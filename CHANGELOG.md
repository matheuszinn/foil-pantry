# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-05-25

### Added
- Implemented Fuzzy Name Sanitization (TASK_10) to strip bracketed tags (e.g. regions, mods, updates, DLCs, and Title IDs) from Switch torrent files for cleaner CLI results and database records.
- Implemented Automated Retry Policy (TASK_10) to handle transient errors during Telegram search/downloads and MinIO uploads.
- Verified end-to-end integration (TASK_08) confirming all acceptance criteria: functional search (AC-1), file persistence in MinIO (AC-2), DB status transitions to `METADATA_READY` (AC-3), and database status logging as `FAILED` during simulated storage outages (AC-4).
- Implemented `DownloadFile` method in [client.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/client.go) to support document downloads from Telegram.
- Created unit tests for the `DownloadFile` method parsing in [client_download_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/client_download_test.go).
- Created the application orchestration in [app.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/app/app.go) to manage config loading, GORM DB, MinIO storage and Telegram clients, execute searches, display numbered results, and execute the download-upload pipeline.
- Implemented the command-line argument parser in [main.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/cmd/downloader/main.go).
- Implemented `Search` method on `ClientWrapper` to resolve search bot usernames, send search queries, poll chat history, and parse search results in [search.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/search.go).
- Implemented `parseSearchResult` to validate file extensions and parse torrent metadata, extracting a robust serialized `FileID` for later file downloading.
- Added comprehensive unit tests in [search_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/search_test.go) verifying parsing rules, message sender filtering, and update message ID extraction.
- Implemented `ClientWrapper` around `github.com/gotd/td` in [client.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/client.go) to manage connections, authentication flows, and expose the MTProto client interface.
- Implemented interactive stdin credentials authenticator supporting phone number, 2FA password, and verification code prompts.
- Added comprehensive unit tests in [client_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/telegram/client_test.go) verifying instantiation, automatic session parent directory creation, and interactive prompts fallback.
- Added `github.com/gotd/td` and `go.uber.org/zap` dependencies.
- Created initial directory structure for Milestone 1: `cmd/downloader`, `internal/config`, `internal/db`, `internal/storage`, `internal/telegram`, `internal/app`, and `pkg/switchutil`.
- Created placeholder entrypoint [main.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/cmd/downloader/main.go) in `cmd/downloader` that outputs initialization log messages.
- Created [tasks.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/tasks.md) containing the atomized and sequential task list for Milestone 1.
- Created `implementation_plan.md` detailing the design, architecture decisions, and verification steps.
- Added `gopkg.in/yaml.v3` dependency to `go.mod` for configuration loading.
- Implemented validation logic for all mandatory parameters in [config.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/config/config.go).
- Added comprehensive configuration validation unit tests in [config_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/config/config_test.go).
- Created [config.yaml.example](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/config.yaml.example) at the project root directory.
- Added GORM drivers `gorm.io/driver/mysql` and `gorm.io/driver/sqlite` dependencies to `go.mod`.
- Implemented `Download` database model in [models.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/db/models.go) mapping strictly to the `downloads` table specifications.
- Implemented `InitDB` in [db.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/db/db.go) with MariaDB connection configuration and structured error return instead of panics.
- Created `Repository` interface and `SQLRepository` implementation in [repository.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/db/repository.go) with status verification and context propagation.
- Created repository unit tests in [repository_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/db/repository_test.go) executing schema migration and CRUD validations on an in-memory SQLite database.
- Added `github.com/minio/minio-go/v7` dependency to `go.mod`.
- Created `StorageEngine` interface and `MinIOStorage` struct in [storage.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/storage/storage.go) for MinIO file uploads.
- Created `MockStorage` and interface verification unit tests in [storage_test.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/storage/storage_test.go).

### Changed
- Refactored configuration validation to define and implement a `Validatable` interface across all config subsystems in [config.go](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/internal/config/config.go).
- Updated [AGENTS.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/AGENTS.md) to add Section 5 (Task Generation Protocol) and Section 6 (Architecture Specification Protocol).
- Translated `spec.md` from Portuguese to English.
- Translated `architecture.md` from Portuguese to English.
- Updated [README.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/README.md) Milestone 1 link to point to [spec.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/spec.md).
- Updated [spec.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/spec.md) to set standard completed metadata status strictly to `METADATA_READY` (aligning with `architecture.md`), and added Section 7 (Definition of Done & Nice to Have).
- Added `telegram_file_id` string field to `downloads` table modeling in [architecture.md](file:///Users/matheuszin/Programming/telegram-tinfoil-downloader/architecture.md).

### Removed
- Deleted redundant files: `docs/milestones/milestone-1-architecture.md`, `docs/technical-details.md`, and `docs/milestones/01-telegram-downloader.md`.
