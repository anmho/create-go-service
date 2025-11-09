# Changelog

All notable changes to the `create-go-service` CLI tool will be documented in this file.

## [Unreleased]

### Added
- **Hybrid Configuration System**: Services now use YAML files for non-sensitive config and environment variables for secrets
  - `config.yaml` for server settings, feature flags, etc. (committed to git)
  - `.env` for secrets like JWT keys and database credentials (gitignored)
  - `.env.example` template showing all required environment variables
  - Configuration loading with `gopkg.in/yaml.v3` and `github.com/caarlos0/env/v10`
  - Environment variables can override any YAML setting
  - Secrets are loaded only from environment variables (never from YAML)

- **Configuration Documentation**:
  - Comprehensive `docs/CONFIG.md` guide
  - Examples for local development and production deployment
  - Troubleshooting section
  - Best practices for secret management

- **Template Improvements**:
  - All templates now use `.HasDynamoDB`, `.HasPostgres`, `.HasAuth`, `.HasMetrics` flags
  - Consistent conditional logic across all template files
  - Better error handling in generated services
  - Graceful shutdown in main.go template

- **Generated Service Features**:
  - Environment-aware reflection control (enabled locally, disabled in production)
  - Config path customization via `CONFIG_PATH` environment variable
  - Structured logging with configuration details
  - Docker support with config file copying
  - Hot reload watches `config.yaml` for changes

### Changed
- Updated all templates to use boolean flags instead of string comparisons
- Improved template conditional logic for cleaner generated code
- Enhanced README with configuration section

### Fixed
- Template syntax errors with database type conditionals
- Empty struct generation when no database is selected
- Inconsistent conditional logic across templates

## [0.1.0] - Initial Release

### Added
- Interactive TUI with Bubbletea
- Support for Chi, Huma, and gRPC (ConnectRPC) API frameworks
- DynamoDB and PostgreSQL/Supabase database options
- Metrics, authentication, and hot reload features
- Fly.io deployment configuration
- GitHub Actions workflow for automatic deployment
- Folder-by-feature project structure
- Docker and docker-compose support
- Makefile with common commands

