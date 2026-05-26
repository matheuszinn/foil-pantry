# Milestone 2: Torrent Integration

## Objective
Extend the system to handle `.torrent` files and Magnet links forwarded from Telegram.

## Strategy
1. **Detection:** Bot identifies a `.torrent` file or a Magnet link.
2. **Processing:** 
   - Use `anacrolix/torrent` (native Go) or an external client.
   - Extract the payload (NSP/XCI/NSZ).
3. **Storage:** Move completed downloads to the primary storage defined in Milestone 1.

## Considerations
- **Storage Space:** Torrenting requires double space temporarily (downloading + final storage) unless downloading directly to the final destination.
- **Cleanup:** Delete the `.torrent` metadata after successful download.
