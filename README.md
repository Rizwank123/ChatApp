
# ChatApp (Still IN Progress)

ChatApp is a messaging application designed with support for real-time communication between users. Currently, the authentication part of the backend API is implemented using Go (Golang).

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Environment Variables](#environment-variables)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Features

- User registration and login
- JWT-based authentication

## Technologies Used

- **Backend**: Go (Golang)
- **Framework**: [Echo](https://echo.labstack.com/v4) - High-performance, extensible, minimalist web framework for Go
- **Database**: PostgreSQL
- **Authentication**: JWT (JSON Web Token)
- **Encryption**: bcrypt for password hashing
- **Environment Management**: dotenv

## Installation

### Prerequisites

- Go (Golang) 1.16 or higher
- PostgreSQL
- Make sure you have installed the necessary Go packages:
  - goose
  - wire

  ```bash
  go mod tidy
  ```

## Environment Variables

Docker Compose reads its settings from `.env` in the project root. Start with the complete example configuration:

```sh
cp .env.example .env
```

Set a non-empty `REDIS_PASSWORD` and replace the sample `AUTH_SECRET` and `MINIO_SECRET_KEY` with local development values. Swagger defaults to username `admin` and the `SWAGGER_PASSWORD` shown in `.env.example`; change it for your environment. `MINIO_ENDPOINT` is the API container's S3 endpoint; `MINIO_PUBLIC_ENDPOINT` is used in presigned image URLs and must be reachable by the browser. The bucket remains private: use the signed `media.url` returned by the API. Opening `http://localhost:8333/` only reaches the S3 API root, which can return `AccessDenied`. Local frontend builds allow HTTP/WS; production frontend builds should use HTTPS/WSS and set `FRONTEND_ALLOW_INSECURE_API=false`. See `.env.example` for all supported settings.

## Running the Application

### With Docker Compose

Docker with the Compose v2 plugin is required. Compose starts the API, PostgreSQL, Redis, object storage, and runs database migrations.

1. Create and configure `.env` as described above.
2. Build and start the services:

   ```bash
   make docker-up
   ```

   This runs `docker compose up --build` and keeps the logs in the foreground. With the example settings, the frontend is at `http://localhost:8081`, the API is at `http://localhost:7700`, and Swagger login is at `http://localhost:7700/login`. PostgreSQL is published on port `5433` and Redis on port `6380`; the API uses the services' internal Compose ports. API health endpoints are `/health/live` and `/health/ready`.
3. Stop the services when finished:

   ```bash
   make docker-down
   ```

   This runs `docker compose down`. Persistent database, Redis, and object storage data remain in Docker volumes.

You can also run the Compose command directly with `docker compose up --build`.

### Running the Go API directly

For local Go development, start PostgreSQL, Redis, and the configured object storage first, apply the database migrations, then run:

```bash
go run ./cmd
```

## API Endpoints

### Authentication

| Method | Endpoint                   | Description                         | Auth Required |
|--------|-----------------------------|-------------------------------------|---------------|
| POST   | `/api/v1/users`             | Register a new user                 | No            |
| POST   | `/api/v1/users/login`       | User login                          | No            |
| GET    | `/api/v1/users/{username}`  | Get user details by username         | Yes           |

## Contributing

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch-name`).
3. Make your changes and commit them (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch-name`).
5. Open a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
