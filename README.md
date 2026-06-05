# Caos — Content-Addressed Object Store

Caos is a content-addressed object store. Objects are stored and retrieved
by their SHA-256 hash (the "address"). Objects can be tagged with
key/value metadata, aliased with human-readable names, and bundled into
directory-like "paths" for HTTP serving.

## Architecture

```
┌──────────────────────────────────────────┐
│            caos server (Go)              │
│                                          │
│  ┌──────────┐  ┌──────────────────────┐  │
│  │  Server   │  │  apiImpl             │  │
│  │  (routes) │──│  (handler methods)   │  │
│  └──────────┘  │  ┌─────┐ ┌─────────┐ │  │
│                │  │Meta │ │  Data   │ │  │
│                │  │store│ │  store  │ │  │
│                │  │(Peb)│ │  (FS)   │ │  │
│                │  └─────┘ └─────────┘ │  │
│                └──────────────────────┘  │
└──────────────────────────────────────────┘
```

### Store layer

- **Meta store** (`server/store/meta.go`): Pebble-backed key-value store for
  addresses, tags, and name aliases. Key schema:

  ```
  a/<addr>       → ""           address existence marker
  t/<addr>/<tag> → <value>      tag value
  n/<name>       → <addr>       name→address mapping
  ```

- **Data store** (`server/store/data.go`): Filesystem-backed blob storage.
  Blobs are named by their SHA-256 hex string. Writes use temp-file-then-rename
  for atomicity.

### Code generation

The HTTP routing and type definitions are generated from `openapi.yaml` using
`oapi-codegen`:

```
make codegen   # runs: go tool oapi-codegen -config openapi-codegen.yaml openapi.yaml
```

Generated output: `server/openapi.gen.go` — `ServerInterface` + `HandlerWithOptions`
using Go 1.22+ `net/http` routing patterns.

## CLI Usage

```
caos serve [--port <port>] [--root <dir>]    Start the HTTP server
caos add <path>                               Upload a file, print SHA-256
caos addr <partial>                           Resolve a partial address
caos tag <addr> <key> [value] [--delete]      Get/set/delete a tag
caos name <name> [<addr>]                     Get/set a name alias
caos get <addr> [--output <file>]             Download data by address
caos push <path...>                           Upload directory tree as path object
caos pull <addr...> [--output <dir>]          Pull path object files to disk
```

### Server

```sh
caos serve --port 31923 --root /tmp/caos-root
```

The server stores meta data in `<root>/meta` (Pebble) and blobs in `<root>/data`
(filesystem). Default root is `/tmp/caos-<pid>`.

### Client commands

All client commands connect to a running server (default `http://localhost:31923`,
overridable with `--server`):

```sh
# Upload a file
caos add myfile.txt
# → b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

# Download by address
caos get b94d27b9 > myfile.txt

# Resolve a partial address
caos addr b94d27
# → b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

# Set a type tag
caos tag b94d27b9 type text/html

# Create a name alias
caos name myfile b94d27b9
# Resolve it
caos name myfile
# → b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

```

### Push & pull

```sh
# Push a directory tree as a path object
caos push mydir/
# → http://localhost:31923/path/f622794b (2 files)

# Pull a path object's files to disk
caos pull f622794b --output ./out
# → f622794b -> ./out
```

## API

All endpoints are defined in `openapi.yaml`. The Go server implements:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Root redirect to home path |
| GET | `/addr/{addr}` | Resolve partial/full address (200/300/404) |
| POST | `/data` | Upload data, returns SHA-256 hash |
| GET | `/data/{addr}` | Retrieve data by address |
| GET | `/tags/{addr}` | Get all tags as JSON |
| GET | `/tags/{addr}/{tag}` | Get a single tag value |
| PUT | `/tags/{addr}/{tag}` | Set a tag |
| DELETE | `/tags/{addr}/{tag}` | Delete a tag |
| GET | `/name/{name}` | Resolve name to address |
| POST | `/name/{name}` | Set a name alias |
| GET | `/path/{addr}/` | Autoindex HTML listing |
| GET | `/path/{addr}/{name}` | Serve a file from a path |
| CLI | `caos push <path>` | Upload directory tree, create path object |
| CLI | `caos pull <addr>` | Download path object files to disk |

## Development

```sh
make build            # codegen + build binary
make codegen          # regenerate openapi.gen.go
make test             # run all tests
make test-verbose     # run tests with -race
```

The integration test suite (`internal/test/`) starts a real server, runs 31
test cases across 6 phases, and cleans up. Tests are ordered by dependency:
data upload → address resolution → tags → names → paths → push/pull.
