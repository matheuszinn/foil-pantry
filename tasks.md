# Milestone 1 Task List

- [X] **TASK_01 - Initialize Go Module and Project Structure**
  - **Context:** Creates `go.mod`, `go.sum`, directories under `cmd/` and `internal/`, and a dummy `main.go`.
  - **Implementation Instructions:**
    1. Initialize the Go module: `go mod init github.com/matheuszin/telegram-tinfoil-downloader`.
    2. Create the following directories:
       - `cmd/downloader`
       - `internal/config`
       - `internal/db`
       - `internal/storage`
       - `internal/telegram`
       - `internal/app`
       - `pkg/switchutil`
    3. Create a basic placeholder `main.go` inside `cmd/downloader/main.go` that prints a start message.
    4. Run `go mod tidy` to verify initial setup.
  - **Dependencies:** None.
  - **Validation Criteria:** Running `go run cmd/downloader/main.go` prints the start message without any compilation issues.

- [X] **TASK_02 - Implement Config Loader and YAML Parsing**
  - **Context:** Creates `internal/config/config.go`, `config.yaml.example`.
  - **Implementation Instructions:**
    1. Define structures for the application configuration matching:
       - `database`: host, port, user, password, dbname
       - `minio`: endpoint, access_key, secret_key, bucket, use_ssl
       - `telegram`: api_id, api_hash, phone, session_path, search_bot
    2. Write a `LoadConfig(path string) (*Config, error)` function utilizing `gopkg.in/yaml.v3`.
    3. Implement validation logic ensuring all mandatory parameters are set, failing immediately with clean error messages if validation fails.
    4. Create `config.yaml.example` with placeholder credentials.
    5. Write unit tests in `internal/config/config_test.go` verifying correct loading and validation.
  - **Dependencies:** TASK_01
  - **Validation Criteria:** Running `go test ./internal/config/...` passes successfully.

- [X] **TASK_03 - Implement Database Modeling and MariaDB Connection**
  - **Context:** Creates `internal/db/models.go`, `internal/db/db.go`.
  - **Implementation Instructions:**
    1. Define GORM model `Download` mapping to `downloads` table structure:
       - `id` (uint, PK)
       - `title_name` (string)
       - `telegram_msg_id` (int)
       - `telegram_file_id` (string)
       - `storage_path` (string)
       - `status` (string) (Must only be `PENDING`, `METADATA_READY`, or `FAILED`)
       - `size_bytes` (int64)
       - `created_at` (time.Time)
       - `updated_at` (time.Time)
    2. Implement `InitDB(cfg *config.Config) (*gorm.DB, error)` to connect to MariaDB and execute GORM's `AutoMigrate`.
    3. Implement a repository structure to:
       - Create a new download record.
       - Update download record status and storage path.
    4. Add unit tests for migrations and GORM operations (using an in-memory SQLite database to decouple unit tests from a running MariaDB instance).
  - **Dependencies:** TASK_02
  - **Validation Criteria:** Running `go test ./internal/db/...` passes successfully.

- [X] **TASK_04 - Implement Storage Abstraction and MinIO Client**
  - **Context:** Creates `internal/storage/storage.go`.
  - **Implementation Instructions:**
    1. Define `StorageEngine` interface with method `UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error`.
    2. Implement `MinIOStorage` struct satisfying the interface using `github.com/minio/minio-go/v7`.
    3. Implement client initializer `NewMinIOStorage(cfg *config.Config) (*MinIOStorage, error)` which checks bucket existence and creates the bucket if it does not exist.
    4. Write interface verification tests/mocks in `internal/storage/storage_test.go`.
  - **Dependencies:** TASK_02
  - **Validation Criteria:** Running `go test ./internal/storage/...` passes successfully.

- [X] **TASK_05 - Implement Telegram Client Wrapper (gotd Connection & Auth)**
  - **Context:** Creates `internal/telegram/client.go`.
  - **Implementation Instructions:**
    1. Set up a wrapper structure around the `github.com/gotd/td` client.
    2. Implement an authentication flow:
       - If no session storage is found at the configured path, prompt the user for their phone number and the verification code sent by Telegram via stdin (interactive login).
       - Persist the session locally to disk using gotd's session storage.
    3. Integrate structured logging (using zap logger) to log connection events.
  - **Dependencies:** TASK_02
  - **Validation Criteria:** Ensure the code compiles. Write tests in `internal/telegram/client_test.go` or verify with mock servers.

- [X] **TASK_06 - Implement Telegram Bot Messaging and Search Parser**
  - **Context:** Creates `internal/telegram/search.go`.
  - **Implementation Instructions:**
    1. Implement a method `Search(ctx context.Context, query string) ([]SearchResult, error)` to:
       - Resolve the username of the configured Search Bot.
       - Send a message text containing the query to the bot.
       - Wait for responses and parse Name, Size, and Date from incoming messages.
       - Extract the unique Telegram message/file reference (`file_id` or checksum) to download it later.
    2. Define `SearchResult` struct to keep metadata in memory:
       - `ID` (int) - sequential index.
       - `Name` (string)
       - `Size` (int64)
       - `Date` (time.Time)
       - `MsgID` (int)
       - `FileID` (string)
    3. Validate that target files have the `.torrent` extension, ignoring other file formats.
    4. Write tests with mock message contents to verify search parsing logic.
  - **Dependencies:** TASK_05
  - **Validation Criteria:** Running `go test ./internal/telegram/...` passes successfully.

- [X] **TASK_07 - Implement CLI Orchestration**
  - **Context:** Modifies `cmd/downloader/main.go`, creates `internal/app/app.go`.
  - **Implementation Instructions:**
    1. In `internal/app/app.go`, implement the command handler orchestration:
       - Initialize Zap structured logger.
       - Initialize configuration, database, MinIO storage engine, and Telegram client.
       - Execute search and print a numbered results list (Name, Size, Date).
       - Prompt the user to input the ID from the list via stdin.
       - When selected, download the target file from Telegram.
       - Add a record to the database `downloads` table with status `PENDING`.
       - Upload the file to Minio (`torrents/<filename>.torrent`).
       - If upload succeeds, update the database status to `METADATA_READY` and store `storage_path`.
       - If upload/operation fails, keep status as `FAILED` (satisfying AC-4: do not mark status as ready).
    2. Implement SIGINT/SIGTERM listener to close database connections and Telegram session cleanly.
    3. Set up the CLI commands parser using `flag` or custom arguments routing in `cmd/downloader/main.go`.
  - **Dependencies:** TASK_03, TASK_04, TASK_06
  - **Validation Criteria:** Running `go build -o downloader ./cmd/downloader` builds the binary successfully.

- [X] **TASK_08 - End-to-End Integration and Manual Verification**
  - **Context:** Modifies `CHANGELOG.md`.
  - **Implementation Instructions:**
    1. Build the CLI binary.
    2. Spin up test MariaDB and MinIO containers using a local environment setup.
    3. Configure credentials in `config.yaml`.
    4. Execute `./downloader search "Zelda"` and select an item.
    5. Verify the `.torrent` file exists in the MinIO bucket.
    6. Verify a row is recorded in `downloads` with status `METADATA_READY` and the correct storage path.
    7. Simulate failure (e.g., stop MinIO service) and confirm the database status does not set to `METADATA_READY`.
    8. Document results and update `CHANGELOG.md`.
  - **Dependencies:** TASK_07
  - **Validation Criteria:** All Milestone 1 Acceptance Criteria (AC-1, AC-2, AC-3, AC-4) are verified and logged in the CHANGELOG.

- [x] **TASK_09 - Implement Docker and Docker Compose Support**
  - **Context:** Creates `Dockerfile`, `docker-compose.yml`, and `config.docker.yaml`.
  - **Implementation Instructions:**
    1. Create a multi-stage `Dockerfile` with Go builder and minimal alpine runtime including CA certificates.
    2. Create `docker-compose.yml` defining `mariadb`, `minio`, and `downloader` services with data persistence volumes and stdin/tty setup.
    3. Create `config.docker.yaml` with hosts configured to resolve container names `mariadb` and `minio`.
  - **Dependencies:** TASK_01, TASK_02
  - **Validation Criteria:** Building with `docker compose build` succeeds, and dependencies can be spun up via `docker compose up -d mariadb minio`.

- [X] **TASK_10 - Implement Fuzzy Name Sanitizer and Retry Policy**
  - **Context:** Creates `pkg/switchutil/sanitizer.go`, `pkg/switchutil/sanitizer_test.go`, and modifies `internal/app/app.go`.
  - **Implementation Instructions:**
    1. Implement bracketed tags removal (`CleanTitle`) in `pkg/switchutil/sanitizer.go`.
    2. Write comprehensive unit tests in `pkg/switchutil/sanitizer_test.go`.
    3. Integrate into `internal/app/app.go` to sanitize search list outputs and database `title_name` records.
    4. Implement retry mechanism wrapper and apply it to Telegram and MinIO operations.
  - **Dependencies:** TASK_07
  - **Validation Criteria:** Sanitizer tests pass and the search results show cleaned titles.

# Milestone 2 Task List

- [ ] **TASK_11 - Extend Config and DB Schema for Torrent Integration**
  - **Context:** Modifies `internal/config/config.go`, `internal/db/models.go`, `internal/db/repository.go`.
  - **Implementation Instructions:**
    1. Update the `Config` structure in `internal/config/config.go` to add a new `Torrent` config section with a `TempDir` field (defaulting to `downloads/tmp` if empty).
    2. Add `payload_path` (string) to GORM model `Download` in `internal/db/models.go`.
    3. Update `DownloadStatus` enum in `internal/db/models.go` with statuses `DOWNLOADING` and `COMPLETED`.
    4. Update `validateStatus` helper to support `DOWNLOADING` and `COMPLETED`.
    5. Implement `GetByStatus(ctx context.Context, status DownloadStatus) ([]Download, error)` in `Repository` interface and its SQL implementation in `internal/db/repository.go`.
    6. Implement `UpdatePayloadPathAndStatus(ctx context.Context, id uint, status DownloadStatus, payloadPath string) error` in `Repository` and its SQL implementation.
  - **Dependencies:** TASK_10
  - **Validation Criteria:** Running `go test ./internal/db/...` and `go test ./internal/config/...` passes.

- [ ] **TASK_12 - Extend Storage Engine with Download Capability**
  - **Context:** Modifies `internal/storage/storage.go` and `internal/storage/storage_test.go`.
  - **Implementation Instructions:**
    1. Add `DownloadFile(ctx context.Context, objectName string, writer io.Writer) error` to the `StorageEngine` interface in `internal/storage/storage.go`.
    2. Implement `DownloadFile` on `MinIOStorage` using the MinIO Go SDK's `GetObject` function, reading the stream and writing it to the input writer.
    3. Update `MockStorage` in `internal/storage/storage_test.go` and write a unit test to verify that `DownloadFile` functions correctly.
  - **Dependencies:** TASK_11
  - **Validation Criteria:** Running `go test ./internal/storage/...` passes.

- [ ] **TASK_13 - Implement Native Go Torrent Downloader Client**
  - **Context:** Creates `internal/torrent/torrent.go`, `internal/torrent/torrent_test.go`.
  - **Implementation Instructions:**
    1. Add `github.com/anacrolix/torrent` and `github.com/schollz/progressbar/v3` as dependencies.
    2. Create `internal/torrent/torrent.go`.
    3. Implement `TorrentDownloader` struct or functions that:
       - Instantiates a torrent client using a configuration structure.
       - Configures download directory.
       - Adds torrent from memory bytes or a `.torrent` file reader.
       - Downloads files to the local temp directory.
       - Exposes progress stats (downloaded bytes, total bytes) continuously or accepts a progress updater callback.
       - Shuts down the client cleanly.
    4. Write mock/unit tests in `internal/torrent/torrent_test.go` to verify the torrent manager interface.
  - **Dependencies:** TASK_11
  - **Validation Criteria:** Running `go test ./internal/torrent/...` passes.

- [ ] **TASK_14 - Implement Torrent Processing Orchestration and CLI Progress Bar**
  - **Context:** Modifies `internal/app/app.go`.
  - **Implementation Instructions:**
    1. In `internal/app/app.go`, implement the function `ProcessTorrents(ctx context.Context, configPath string, verbose bool) error`:
       - Initialize Logger, Config, DB Repository, StorageEngine, and Torrent Client.
       - Retrieve all downloads with status `METADATA_READY`.
       - For each download:
         a. Transition status to `DOWNLOADING` in the database.
         b. Download `.torrent` file from MinIO storage into memory.
         c. Start torrent download via the client to the temp folder.
         d. Render a dynamic visual progress bar on the CLI using `github.com/schollz/progressbar/v3`. Do not print standard logs that would mess up the progress bar.
         e. Once download completes, scan the downloaded files to filter files matching Nintendo Switch extensions (`.nsp`, `.xci`, `.nsz`).
         f. Upload the matching files to MinIO under prefix `games/` or similar.
         g. Clean up the downloaded local temporary files.
         h. Set the status to `COMPLETED` and update `payload_path` in the database with the MinIO upload paths (JSON-serialized or comma-separated if multiple files).
         i. Handle failures gracefully: if any download or upload step fails, set database status to `FAILED`.
  - **Dependencies:** TASK_12, TASK_13
  - **Validation Criteria:** The code compiles successfully without errors.

- [ ] **TASK_15 - Extend CLI Commands Parser to Support Processing Torrents**
  - **Context:** Modifies `cmd/downloader/main.go`.
  - **Implementation Instructions:**
    1. Refactor argument parsing in `cmd/downloader/main.go` to support two commands:
       - `search <query>`: executes existing search and download metadata pipeline (from Milestone 1).
       - `process-torrents`: executes the new `ProcessTorrents` orchestration.
    2. Update CLI usage text.
  - Dependencies: TASK_14
  - **Validation Criteria:** Running `go build -o downloader ./cmd/downloader` builds the binary successfully, and running `./downloader process-torrents` starts execution.

- [ ] **TASK_16 - Verify Milestone 2 End-to-End**
  - **Context:** Modifies `CHANGELOG.md`.
  - **Implementation Instructions:**
    1. Build the updated CLI binary.
    2. Run search and capture torrent metadata for a test item using `./downloader search "..."`. Verify it is stored in MinIO and DB status is `METADATA_READY`.
    3. Run `./downloader process-torrents` to start the download.
    4. Verify that the CLI displays real-time progress.
    5. Verify that the final game payload file is uploaded to MinIO.
    6. Verify that the database status is updated to `COMPLETED` and `payload_path` is correctly populated.
    7. Verify that temporary files on the local disk are completely cleaned up.
    8. Document results in `CHANGELOG.md`.
  - **Dependencies:** TASK_15
  - **Validation Criteria:** All Milestone 2 requirements verified and logged in the CHANGELOG.


