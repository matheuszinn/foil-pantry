# Milestone 4: Switch Integration

## Objective
Configure the Nintendo Switch console to use the custom server as a source in Tinfoil.

## Steps

### 1. Network Preparation
- Ensure the Switch and the Server are on the same network, or the Server is publicly accessible.
- Note the Server's IP address or Domain.

### 2. Tinfoil Configuration
1. Open **Tinfoil** on the Switch.
2. Navigate to **File Browser**.
3. Press `X` to add a new location.
4. Set Protocol to `http` or `https`.
5. Enter the **Host** (IP/Domain) and **Port**.
6. (Optional) Set a **Title** like "My Private Shop".
7. Save and refresh.

### 3. Verification
- The games downloaded via the Telegram bot (Milestone 1 & 2) should now appear in the "New Games" tab of Tinfoil.
- Test a download to ensure the server handles requests correctly.
