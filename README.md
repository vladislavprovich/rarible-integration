# Rarible Service

This project is a service integration for Rarible API, providing endpoints to query NFT ownership and traits with rarity.

## Running Locally

### Prerequisites
- Go 1.18 or higher
- Docker (optional, for containerized run)
- Make
- Task (https://taskfile.dev/#/installation)

### Using Makefile

To build and run the service locally on Unix-like shells:

```bash
make build
make run
```

- `make build` compiles the project.
- `make run` runs the compiled binary.

### Using Taskfile

You can also use Taskfile commands on Unix-like shells:

```bash
task build
task run
```

- `task build` compiles the project.
- `task run` runs the compiled binary.

### Running with Docker

To build and run the service using Docker on Unix-like shells:

```bash
docker build -t rarible-service .
docker run -p 8080:8080 --env-file ./.env rarible-service
```

This will expose the service on port 8080.

### Using Docker Compose

You can also use Docker Compose on Unix-like shells:

```bash
docker-compose up --build
```

This will build and start the container as defined in `docker-compose.yaml`.

## API Endpoints

Base URL: `http://localhost:8080/api/v1/rarible`

- `GET /ownership` - Get NFT ownership by ID
- `POST /rarities` - Query traits with rarity

## Example Requests

### Get Ownership by ID

On Unix-like shells:

```bash
curl -X GET http://localhost:8080/api/v1/rarible/ownership -d '{"id":"some-nft-id"}' -H "Content-Type: application/json"
```

### Query Traits with Rarity

On Unix-like shells:

```bash
curl -X POST http://localhost:8080/api/v1/rarible/rarities -d '{"traits": [...]}' -H "Content-Type: application/json"
```

Replace the JSON bodies with appropriate request data.

## Configuration

Configuration is loaded from environment variables or config files as per the project setup.

## Logging

The service uses structured logging with slog.

---

For more details, refer to the source code and configuration files.
