# Chariot

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

2. Mongo Migration
```bash
  $ make migrate-mongo-create [args] # args e.g: migrate-mongo-create file-table
  $ make migrate-mongo [args] # args e.g: migrate-mongo up
```

### MySQL Replication Setup
1. Run setup
```bash
  $ ./development/mysql/replication.sh
```

### MongoDB Replication Setup
1. Generate keyFile (if necessary)
```bash
  $ cd /development/mongo
  $ openssl rand -base64 741 > mongodb.key
  $ chmod 400 mongodb.key
```

2. Setting local hosts
- Window
C:\Windows\System32\drivers\etc\hosts
```md
  127.0.0.1 mongo-db-1
  127.0.0.1 mongo-db-2
  127.0.0.1 mongo-db-3
```

- Linux
\etc\hosts
```md
  127.0.0.1 mongo-db-1
  127.0.0.1 mongo-db-2
  127.0.0.1 mongo-db-3
```

3. Run setup
```bash
  $ ./development/mongo/replication.sh
```

## Todo
1. Admin: CreateAuthClient
2. Admin: ReadAuthClient
3. Admin: UpdateAuthClient
4. Admin: SearchAuthClients
5. Admin: CreateBarrel
6. Admin: ReadBarrel
7. Admin: UpdateBarrel
8. Admin: SearchBarrels

9. Client: UploadFile
- *file
- *barrels (hippo, min: 1, max: 10, order define priority)
- *visibility (public, protected)
- metadata (min: 0, max: 30, key-value pairs)
-> return `file_id`, `uploaded_at` for protected file
-> return `file_id`, `uploaded_at` and `file_url` for public file
10. Client: Secure UploadFile (using auth)
- secure upload (e.g: presigned url, upload session, etc)
11. Client: RetrieveFile
- auto failover
12. Client: File access control (visibility, secret meta: user_id)

13. Admin: ReadFile
14. Admin: SearchFiles

15. Client: Upload rule (size, resolution, extension)
- rule is required
- rule may have no attribute (free rule)
- rule may have multiple attribute
- if rule have multiple attribute than it's mean we're matching at least one rule (or clause)

16. Admin: DeleteFile

17. Add barrel provider: (aws-s3)
18. Add barrel provider: (g-storage)
19. Add barrel provider: (alicloud-oss)

20. Admin: dashboard monitoring
- data exporter: CollectMetris
- prometheus (rest exporter)
- grafana

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

## Issue
No issue right now

## Note
1. Make sure X-Correlation-Id is in a string data type and not greater than 128 char length
