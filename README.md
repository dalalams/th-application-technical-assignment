# th-application-technical-assignment

A multi-component/service system to manage and discover content.
## Features

- **Content Management**: Create and manage series, episodes, and categories
- **Content Import**: Import content from YouTube and other sources
- **Search**: Full-text search across series and episodes using OpenSearch
- **File Storage**: MinIO integration for file uploads and management
- **Task Processing**: Asynchronous task processing with Redis and Asynq

## Architecture

- **CMS API**: RESTful API for content management (`cmd/cms`)
- **Discovery API**: Search API for content discovery (`cmd/discovery`)
- **Importer Worker**: Processes content import tasks (`cmd/workers/importer`)
- **Indexer Worker**: Handles search indexing tasks (`cmd/workers/indexer`)
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Search**: OpenSearch for full-text search
- **Storage**: MinIO for file storage
- **Queue**: Redis + Asynq for task processing

## Quickstart
### Prerequisites
- Docker and Docker Compose
- Go 1.24+
- gnu make 3.8+
### 1. Clone and Setup
```bash
git clone <repository-url>
cd th-application-technical-assignment
```

### 2. Environment Configuration
Copy the example environment file and configure:
```bash
cp .env.example .env
```

### 4. Start Services
```bash
# start all services
make docker-up
# stop all
make docker-down
```

### 4. Verify Setup
```bash
# check cms api
curl http://localhost:3000/health
# check discovery api  
curl http://localhost:4000/health
```

### 5. Run Tests
```bash
make test
```

## API Endpoints
### CMS API (Port 3000)
- `POST /series` - create series
- `GET /series` - list series
- `POST /series/{id}/episodes` - create episode
- `POST /import` - import content
- `POST /upload/url` - get upload url
**API Documentation**: http://localhost:3000/swagger/index.html
### Discovery API (Port 4000)
- `GET /search/series` - search series
- `GET /search/episodes` - search episodes
**API Documentation**: http://localhost:4000/swagger/index.html

## Development
### Project Structure
```
├── cmd/                   # application binaries
│   ├── cms/               # cms api server
│   ├── discovery/         # discovery api server
│   └── workers/           # background workers
├── internal/              # private application code
├── pkg/                   # public packages
├── migrations/            # database migrations
└── sqlc/                  # generated sql code
```

### Key Commands
```bash
# run tests
make test

# run linter
make lint

# build all binaries
make build
```

### Adding New Content Sources
1. Implement `Importer` interface in `pkg/importer/`
2. Register in `pkg/importer/importers.go`
3. Add validation in `pkg/api/cms/v1/import_types.go`