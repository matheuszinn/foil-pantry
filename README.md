# Telegram Tinfoil Downloader

A Go-based automated pipeline to fetch Nintendo Switch ROMs from Telegram (Direct or Torrent) and serve them directly to a Nintendo Switch via Tinfoil.

## Project Vision
The goal is to bridge the gap between Telegram-based ROM sharing and the ease of use of a Tinfoil "Shop". The project is divided into logical milestones to ensure stable delivery.

## Roadmap

### [Milestone 1: Telegram Downloader (Direct Files)](spec.md)
- **Status:** Planning
- **Goal:** Fetch NSP/XCI/NSZ files directly from Telegram messages/forwards.
- **Tech:** Go, `gotd` (MTProto), Local/MinIO Storage.

### [Milestone 2: Torrent Integration](docs/milestones/02-torrent-integration.md)
- **Status:** Planning
- **Goal:** Handle `.torrent` files and Magnet links forwarded from Telegram.
- **Tech:** `anacrolix/torrent` or Transmission RPC.

### [Milestone 3: Tinfoil Server (The Shop)](docs/milestones/03-tinfoil-server.md)
- **Status:** Planning
- **Goal:** Serve files and generate a Tinfoil-compatible `index.json`.

### [Milestone 4: Switch Integration](docs/milestones/04-switch-integration.md)
- **Status:** Planning
- **Goal:** Connect the physical Switch to the server via Tinfoil.

## Getting Started
*Coming soon as Milestone 1 implementation begins.*
