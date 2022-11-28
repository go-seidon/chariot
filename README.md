# Chariot

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=go-seidon_chariot&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=go-seidon_chariot)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=go-seidon_chariot&metric=coverage)](https://sonarcloud.io/summary/new_code?id=go-seidon_chariot)

Storage aggregator, managing multiple files from various storage provider

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
  $ make run-restapp
  $ make build-restapp
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
7. Client: Secure Access
- GenerateAccessToken
8. Client: Secure RetrieveFile 
- secure retrieval (file_url on upload file result)
- access duration
- add file config (access_host)
9. Client: Secure UploadFile
- secure upload (e.g: presigned url, upload session, etc)

10. Client: Upload rule (size, resolution, extension)
- rule is required
- rule may have no attribute (free rule)
- rule may have multiple attribute
- if rule have multiple attribute than it's mean we're matching at least one rule (or clause)
11. Daemon: ProceedPendingUpload
12. Admin: DeleteFile

13. Admin: dashboard monitoring
- data exporter: CollectMetris
- prometheus (rest exporter)
- grafana
14. Devs: Enhancement
- Override default error handler (echo router)
- Add request logging
- Unit test: app NewDefaultConfig
- Unit test: storage multipart test

## Nice to have
1. File backup
2. SDK (golang, js, php)
3. Middleware (mux, fiber, echo, gin)
4. Repository provider (mongo, postgres)
5. Retrieve image
- Image manipulation capability (width, height, compression)
6. Caching support
7. Enhance Rule: mimetype

## Issue
1. Gorm not inserting has many association, issue since gorm@v1.22.5 [ref](https://github.com/go-gorm/gorm/issues/5754). Current solution is to use gorm@v1.22.4, mysql@v1.2.1, dbresolver@v1.1.0

## Note
1. Make sure X-Correlation-Id is in a string data type and not greater than 128 char length
