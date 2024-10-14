
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

Create a `.env` file with the following configuration:

```bash
APP_NAME=chat_app
APP_ENV=development
APP_PORT=7700

DB_HOST=localhost
DB_PORT=5433
DB_USERNAME=db_user_name
DB_PASSWORD=db_password
DB_DATABASE_NAME=db_name

SWAGGER_HOST_URL=api_url
SWAGGER_HOST_SCHEME=https
SWAGGER_USERNAME=user_name
SWAGGER_PASSWORD=password

AUTH_SECRET=secret
AUTH_EXPIRY_PERIOD=90
```

## Running the Application

1. Ensure PostgreSQL is running and the necessary environment variables are set.
2. Run the application using:

   ```bash
   go run main.go
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
