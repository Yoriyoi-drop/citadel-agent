# Changelog

All notable changes to Citadel Agent will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OAuth 2.0 authentication with GitHub and Google
- Docker Compose configuration for development and production
- OpenAPI/Swagger documentation for API endpoints
- GitHub Actions CI/CD pipelines
- Comprehensive architecture documentation
- Unit tests for core components
- Project structure documentation
- Contribution guidelines
- Security policy
- Code of conduct
- Makefile for common tasks
- Database migration and seeding tools
- Advanced CLI tools
- Terminal login interface with GitHub/Google options

### Changed
- Improved project directory structure
- Enhanced security architecture with sandboxing
- Updated README with comprehensive documentation
- Improved workflow engine with better node execution
- Refined API endpoints with better error handling
- Enhanced documentation with architecture diagrams

### Deprecated
- Direct environment variable setting without .env files

### Removed
- Hardcoded configuration values

### Fixed
- Various bugs in workflow execution
- Security vulnerabilities in authentication
- Performance issues in database queries

## [0.1.0] - 2023-11-XX

### Added
- Initial release of Citadel Agent
- Core workflow engine
- Multi-language runtime support (Go, Python, JavaScript, etc.)
- Node-based workflow system
- Basic authentication and authorization
- REST API endpoints
- Frontend dashboard
- Worker and scheduler services
- Database integration with PostgreSQL
- Redis caching and queuing
- Basic documentation