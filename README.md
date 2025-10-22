# gapstack

Simple Go service that exposes the Transactions API backed by MySQL.

### Prerequisites
- Go 1.25+
- Docker and Docker Compose (optional but recommended)

### Environment variables
The app reads its database configuration from environment variables (a `.env` file in the project root is supported):

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `3306`)
- `DB_USER` (required)
- `DB_PASSWORD` (required)
- `DB_NAME` (default: `mydatabase`)
- `DB_MAX_OPEN_CONNS` (default: `25`)
- `DB_MAX_IDLE_CONNS` (default: `25`)
- `DB_CONN_MAX_LIFETIME_MINUTES` (default: `5`)

Example `.env`:

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=appuser
DB_PASSWORD=apppass
DB_NAME=transactions_db
```

## Run with Docker Compose (recommended)

This starts MySQL and the app together, and initializes the schema from `db/init.sql`.

```bash
docker compose up --build -d
```

Once running:
- API is available at `http://localhost:8080`
- MySQL is exposed on `localhost:3306` (credentials in `docker-compose.yml`)

To view logs:
```bash
docker compose logs -f go_app
```

To stop:
```bash
docker compose down
```

## Run locally (without Docker)

1) Start a MySQL 8 instance and create a database (matching your `DB_NAME`).

2) Apply schema:
```bash
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/init.sql
```

3) Set environment variables (or create `.env` as above), then run the server:
```bash
go run ./cmd/server
```

Server listens on `:8080`.

## Build and run with Docker (without Compose)

```bash
docker build -t gapstack .
docker run --rm -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=3306 \
  -e DB_USER=appuser \
  -e DB_PASSWORD=apppass \
  -e DB_NAME=transactions_db \
  gapstack
```

Note: adjust `DB_*` to point at a reachable MySQL instance.

## API

Base URL: `http://localhost:8080`

- Create transaction
  - `POST /transactions`
  - Body:
    ```json
    {        
      "amount": 100.50,
      "currency": "USD",
      "sender": "Alice",
      "receiver": "Bob"
    }
    ```
  - Notes: `status` defaults to `pending`.

- List transactions
  - `GET /transactions?page=1&page_size=10`

- Get a transaction
  - `GET /transactions/{id}`

- Update a transaction status
  - `PUT /transactions/{id}`
  - Body:
    ```json
    { "status": "completed" }
    ```
    or
    ```json
    { "status": "failed" }
    ```

## Tests

```bash
go test ./...
```

