# Technical Specification: Telegram Tinfoil Downloader - Milestone 1 (POC)

## 1. PRIMARY OBJECTIVE
Provide a CLI tool that automates searching and capturing metadata (.torrent) of Nintendo Switch games via Telegram, persisting the search state in MariaDB and the physical file in MinIO, serving as the foundation for the complete download pipeline automation.

---

## 2. SYSTEM REQUIREMENTS (EARS/Gherkin)

### Title Search
- **GIVEN** the user provides a search string via CLI
- **WHEN** the system interacts with the search bot on Telegram
- **THEN** the system must return a numbered list of results (Name, Size, Date) extracted from the messages.

### Selection and Metadata Download
- **EVENT:** When the user selects an ID from the generated list.
- **ACTION:** The system must request the file from Telegram, download it to memory/temp, and upload it immediately to the configured bucket in MinIO.

### State Persistence
- **GIVEN** a .torrent file has been successfully uploaded to MinIO
- **THEN** the system must create a record in the `downloads` table of MariaDB with status `METADATA_READY` and the object path in MinIO.

---

## 3. SCOPE BOUNDARIES

### In Scope (Must Do)
- Command line interface (CLI) for search and selection.
- MTProto integration (User Session) to allow interaction with other bots/channels.
- Storage abstraction focused exclusively on MinIO.
- Basic persistence in MariaDB (GORM).
- Validation of allowed extensions: `.torrent` (specific to this Milestone).

### Out of Scope (Explicitly Excluded)
- Downloading actual game content (torrent payload) -> **Milestone 2**.
- Reactive Bot Interface (Telebot) -> Postponed until after CLI stabilization.
- Web Interface or Tinfoil Shop Server -> **Milestone 3**.
- Management of multiple Telegram user sessions.

---

## 4. RULES AND INVARIANTS

1. **Metadata Integrity:** Every file saved in MinIO must have a corresponding entry in MariaDB with the original Telegram `checksum` or `file_id`.
2. **Operation Atomicity:** A download record in the database can only be marked as `METADATA_READY` if the upload to MinIO returns success.
3. **Credential Isolation:** No API keys or database credentials should be hardcoded; mandatory use of `config.yaml`.
4. **Search Uniqueness:** Search results are ephemeral and do not need to be persisted, but the user's choice must be traceable.

---

## 5. ACCEPTANCE CRITERIA

| ID | Criterion | Validation |
|:---|:---|:---|
| AC-1 | Functional Search | `./downloader search "Zelda"` returns at least 1 valid item from Telegram. |
| AC-2 | File Persistence | The selected `.torrent` file appears in the MinIO bucket with the correct name. |
| AC-3 | DB Registration | A new row in the `downloads` table is created with the correct status. |
| AC-4 | Graceful Failure | If MinIO is offline, the system must not mark the download as completed in the DB. |

---

## 6. AMBIGUITIES AND TRADE-OFFS

- **MTProto vs Bot API:** Decided to use **MTProto (`gotd`)**. The system will operate as a **User Session**, using the `API_ID` and `API_HASH` of a user account. This choice is mandatory to allow interaction with other search bots, which is impossible via the standard Bot API.
- **Search Ephemerality:** The results list will be kept only in memory during CLI execution.
- **Authentication:** On the first execution, the system will request the phone number and verification code to generate the local session file.

---

## 7. DEFINITION OF DONE (DoD) & NICE TO HAVE

### General Definition of Done (DoD)
A milestone/task is considered completed only when:
1. **Spec Alignment:** All acceptance criteria (AC) defined in the milestone are met and verified.
2. **Code Quality:** All code is cleanly structured, with structured logging (zap) and context propagation (`context.Context`) for all I/O calls.
3. **Test Coverage:** All new logic (parsers, repositories, configs) has unit tests covering happy and edge cases, aiming for at least 80% coverage.
4. **Environment Integrity:** No hardcoded credentials; all secrets must run via `config.yaml`.
5. **Clean Shutdown:** Application intercepts termination signals (SIGINT, SIGTERM) and terminates cleanly without leaving orphan connections or locked files.

### Nice-to-Have Features (Non-blocking for MVP)
- **Dynamic CLI Progress Bar:** Visual indicator showing progress of the download from Telegram and upload to MinIO.
- **Fuzzy Name Sanitization:** Utilize a parser in `pkg/switchutil` to remove bracketed tags (e.g., `[Update]`, `[DLC]`, region tags) and show cleaner game names in the search results.
- **Automated Retry Policy:** Automatically retry failing uploads or Telegram queries a limited number of times before failing.
