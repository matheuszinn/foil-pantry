# Milestone 3: Tinfoil Server (The Shop)

## Objective
A web server that serves ROM files and a dynamically generated `index.json` (or `shop.json`) that Tinfoil can consume.

## Key Components

### 1. File Server
- **Service:** A standard HTTP/HTTPS server (Go `net/http`).
- **Functionality:** 
  - Serve files from the `storage` abstraction (Local or MinIO).
  - Support Range requests (vital for large file downloads on the Switch).

### 2. Dynamic Index Generator
- **Service:** A logic layer that scans the storage.
- **Output:** A JSON object matching the Tinfoil format:
  ```json
  {
    "files": [
      {
        "url": "http://your-server/game.nsp",
        "size": 1234567
      }
    ],
    "directories": []
  }
  ```
- **Performance:** Cache the index and update it only when new files are added by the Milestone 1/2 downloaders.

### 3. Basic Security
- Optional basic auth or token-based URL parameters if the server is exposed to the internet.
