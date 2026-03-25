# CAOS API Specification - Level 0

## Overview

CAOS (Content Addressed Object Store) is a simple HTTP API for storing and retrieving content-addressed objects. This specification defines Level 0 - the core storage functionality with SHA-256 content addressing.

## Core Concepts

### Content Addressing

Objects are stored by their SHA-256 hash (address). Each address is:
- **64 hexadecimal characters** (full SHA-256)
- Unique and immutable
- Deterministic (same content = same address)

### Tags (Metadata)

Objects can have arbitrary metadata tags. Two special tags are auto-set:
- `size`: Object size in bytes (immutable)
- `type`: Content type (defaults to `application/octet-stream`)

### Content Type Verification

Clients can specify a Content-Type header when adding data. The server must verify the Content-Type is valid and that the data matches the alleged type.

## Base URL

`http://localhost:31923`

## Address Format

All addresses must be 64-character hexadecimal strings representing a SHA-256 hash.

**Examples:**
- `d10b49b4c9a5e4b8a6c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1`
- `abc123...` (64 chars total)

## Endpoints

### 1. Add Data

**POST** `/data`

Stores a new data object and returns its SHA-256 address.

**Request:**
- Method: `POST`
- Headers:
  - `Content-Type`: Optional content type
- Body: Binary data

**Response:**
- Status:
  - `200 OK`: Data stored successfully
  - `400 Bad Request`: Invalid content type or type verification failed
  - `500 Internal Server Error`: Server error
- Headers:
  - `Content-Type`: `text/plain`
- Body: 64-character SHA-256 hex address

**Auto-generated tags:**
- `size`: Size of the uploaded data in bytes
- `type`: Content type from header (or `application/octet-stream` if not provided)

---

### 2. Get Data

**GET** `/data/:addr`

Retrieves data for a given address.

**Path Parameters:**
- `addr`: 64-character SHA-256 hex string

**Request:**
- Method: `GET`

**Response:**
- Status:
  - `200 OK`: Data retrieved successfully
  - `400 Bad Request`: Invalid address format (must be 64 hex characters)
  - `404 Not Found`: No address matches the query
- Headers:
  - `Content-Type`: Content type tag value, or `application/octet-stream` if not set
- Body: Binary data

---

### 3. Delete Data

**DELETE** `/data/:addr`

Deletes the data for a given address.

**Path Parameters:**
- `addr`: 64-character SHA-256 hex string

**Request:**
- Method: `DELETE`

**Response:**
- Status:
  - `204 No Content`: Data deleted successfully (or did not exist)
  - `400 Bad Request`: Invalid address format (must be 64 hex characters)
  - `500 Internal Server Error`: Server error

---

### 4. Head Data

**HEAD** `/data/:addr`

Returns metadata for a given address without the body.

**Path Parameters:**
- `addr`: 64-character SHA-256 hex string

**Request:**
- Method: `HEAD`

**Response:**
- Status:
  - `200 OK`: Address exists
  - `400 Bad Request`: Invalid address format (must be 64 hex characters)
  - `404 Not Found`: No address matches the query
- Headers:
  - `Content-Type`: Content type tag value, or `application/octet-stream` if not set
  - `Content-Length`: Size of the data in bytes
- Body: None

---

## Error Codes

|| Code | Description | When to Use |
||------|-------------|-------------|
|| `200` | Success | Data retrieved successfully |
|| `204` | No Content | Data deleted successfully (or did not exist) |
|| `400` | Bad Request | Invalid address format or invalid content type |
|| `500` | Internal Server Error | Server error |

## Compliance

A CAOS server implementation is Level 0 compliant if it:

1. Implements the required endpoints (POST /data, GET /data/:addr)
2. Uses SHA-256 for content addressing
3. Returns 64-character hexadecimal addresses
4. Validates address format (must be 64 hex characters)
5. Validates Content-Type header (reject unknown types)
6. Performs Content-Type verification when provided
7. Properly handles errors with correct status codes
8. Automatically sets 'size' and 'type' tags
9. Supports binary data

## Version History

- **Level 0** (current): Core storage with SHA-256 addressing and Content-Type verification
- Future levels will add tags, names, paths, references, etc.

## References

- [SHA-256 Hash](https://en.wikipedia.org/wiki/SHA-2)
- [Content Addressable Storage](https://en.wikipedia.org/wiki/Content-addressable_storage)