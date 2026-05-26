# Foil Pantry (Telegram Tinfoil Downloader)

A Go-based automated pipeline to fetch Nintendo Switch ROMs from Telegram (Direct or Torrent) and serve them directly to a Nintendo Switch via Tinfoil.

## Project Vision
The goal is to bridge the gap between Telegram-based ROM sharing and the ease of use of a Tinfoil "Shop". The project is divided into logical milestones to ensure stable delivery.

## Project Progress

- [x] **Milestone 1: Telegram Downloader (Direct Files)**
  - Automated search and metadata extraction from Telegram bots.
  - Integration with MariaDB for tracking and MinIO for storage.
  - Interactive CLI and Docker support.
  - Title sanitization and retry policies.
- [ ] **Milestone 2: Torrent Integration** (In Progress)
  - Native Go torrent client implementation.
  - Automated download of payloads from tracked torrents.
- [ ] **Milestone 3: Tinfoil Server**
- [ ] **Milestone 4: Switch Integration**

## Roadmap

### [Milestone 1: Telegram Downloader (Direct Files)](spec.md)
- **Status:** Completed ✅
- **Goal:** Fetch NSP/XCI/NSZ files directly from Telegram messages/forwards.
- **Tech:** Go, `gotd` (MTProto), MariaDB, MinIO Storage.

### [Milestone 2: Torrent Integration](docs/milestones/02-torrent-integration.md)
- **Status:** In Progress 🚧
- **Goal:** Handle `.torrent` files and Magnet links forwarded from Telegram.
- **Tech:** `anacrolix/torrent`.

### [Milestone 3: Tinfoil Server (The Shop)](docs/milestones/03-tinfoil-server.md)
- **Status:** Planning
- **Goal:** Serve files and generate a Tinfoil-compatible `index.json`.

### [Milestone 4: Switch Integration](docs/milestones/04-switch-integration.md)
- **Status:** Planning
- **Goal:** Connect the physical Switch to the server via Tinfoil.

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Telegram API Credentials ([my.telegram.org](https://my.telegram.org))

### Setup
1. **Clone the repository**:
   ```bash
   git clone https://github.com/matheuszinn/foil-pantry.git
   cd foil-pantry
   ```

2. **Configure environment**:
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your Telegram api_id, api_hash and phone
   ```

3. **Spin up dependencies**:
   ```bash
   docker compose up -d mariadb minio
   ```

4. **Run the downloader**:
   ```bash
   # Run via Docker
   docker compose run --rm downloader search "Zelda"

   # Or run locally (requires Go 1.22+)
   go run cmd/downloader/main.go search "Zelda"
   ```

On the first run, the application will prompt for your Telegram phone number and verification code to establish a secure session.
