# Chariot

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=go-seidon_chariot&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=go-seidon_chariot)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=go-seidon_chariot&metric=coverage)](https://sonarcloud.io/summary/new_code?id=go-seidon_chariot)

Storage provider aggregator, managing multiple files from various location

## Technical Stack
1. Transport layer
- rest
2. Database
- mysql
3. Config
- system environment
- file (config/*.toml and .env)

## How to Run?
### Test
1. Unit test

This particular command should test individual component and run really fast without the need of involving 3rd party dependencies such as database, disk, etc.

```
  $ make test-unit
  $ make test-watch-unit
```

2. Integration test

This particular command should test the integration between component, might run slowly and sometimes need to involving 3rd party dependencies such as database, disk, etc.

```
  $ make test-integration
  $ make test-watch-integration
```

3. Coverage test

This command should run all the test available on this project.

```
  $ make test
  $ make test-coverage
```

### App
1. REST App

```
  $ make run-rest-app
  $ make build-rest-app
```

### Docker
TBA

## Development
### First time setup
1. Copy `.env.example` to `.env`

2. Create docker compose
```bash
  $ docker-compose up -d
```

### Database migration
1. MySQL Migration
```bash
  $ make migrate-mysql-create [args] # args e.g: migrate-mysql-create file-table
  $ make migrate-mysql [args] # args e.g: migrate-mysql up
```

### MySQL Replication Setup
1. Run setup
```bash
  $ ./development/mysql/replication.sh
```

## Todo
1. Devs: Adjustment code structure
- Move mysql repo as repository subfolder
- Remove dash foldering name

2. Client: RetrieveFile
- auto failover
3. Client: GenerateAccessToken
4. Client: RetrieveFile 
- secure access (file_url on upload file result)
- add file config (access_host)
5. Admin: GetFileById
6. Admin: SearchFiles
7. Admin: DeleteFile
8. Daemon: ProceedPendingUpload
9. Client: Secure UploadFile (using auth)
- secure upload (e.g: presigned url, upload session, etc)
10. Client: Upload rule (size, resolution, extension)
- rule is required
- rule may have no attribute (free rule)
- rule may have multiple attribute
- if rule have multiple attribute than it's mean we're matching at least one rule (or clause)
11. Admin: dashboard monitoring
- data exporter: CollectMetris
- prometheus (rest exporter)
- grafana
12. Devs: Enhance rest app
- Override default error handler (echo router)

## Nice to have
1. Custom file access (custom link with certain limitation, e.g: access duration)
2. File backup
3. SDK
- golang
- javascript
- php
4. Middleware
- mux
- fiber
- echo
- gin
5. Repository provider
- mongo
- postgre
6. Retrieve image
- Image manipulation capability (width, height, compression)
7. Caching support
8. Enhance Rule: mimetype
9. Unit test: app NewDefaultConfig

## Issue
No issue right now

## Note
1. Make sure X-Correlation-Id is in a string data type and not greater than 128 char length
