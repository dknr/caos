# CAOS Level 0 Compliance Checker

This repository contains a minimal, pure‑standard‑library Go program that can be used to verify that a CAOS server implements the Level 0 API correctly. The program performs the following sequence of requests against a running server:

1. **POST /data** – uploads a small payload and receives a 64‑character SHA‑256 address.
2. **GET /data/:addr** – downloads the same payload and verifies the body.
3. **HEAD /data/:addr** – checks the `Content‑Type` and `Content‑Length` headers.
4. **DELETE /data/:addr** – deletes the object and expects `204 No Content` (or `200`).
5. **GET /data/:addr** – after deletion, expects `404 Not Found`.

If all steps succeed the program prints
```
All compliance tests passed.
```
and exits with status 0. Any failure results in a non‑zero exit and a message describing the error.

## Building

```bash
go build -o caos-comply ./docs/v0/caos-comply
```

The resulting binary can be invoked directly:

```bash
./caos-comply --url http://localhost:31923
```

The `--url` flag defaults to `http://localhost:31923`, matching the default port used by the `caos serve` command.

## Usage Example

```bash
# Start the server on the default port
./caos serve

# In a separate terminal run the compliance checker
./caos-comply
```

The checker will perform all tests and exit with success.

---

*This program is intentionally free of external dependencies and is suitable for use in CI pipelines or as part of a manual test suite.*
