# Forum

A modern web forum application built with Go, SQLite, and Docker. The forum features user authentication, post management with categories, likes/dislikes system, and post filtering capabilities.

## Features

- User Authentication (Register/Login)
- Post Creation and Management
- Categorized Posts
- Comments System
- Like/Dislike System for Posts and Comments
- Post Filtering by Categories
- Clean and Modern UI with Deep Purple Theme
- Secure Password Encryption
- Session Management with Cookies

## Tech Stack

- Backend: Go
- Database: SQLite3
- Frontend: HTML, CSS, JavaScript (Vanilla)
- Containerization: Docker

## Project Structure

```
forum/
├── cmd/
│   └── main.go
├── internal/
│   ├── auth/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── utils/
├── static/
│   ├── css/
│   ├── js/
│   └── img/
├── templates/
├── tests/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.16 or higher
- Docker
- Docker Compose

## Setup and Running

1. Clone the repository:
```bash
git clone <repository-url>
cd forum
```

2. Build and run with Docker:
```bash
docker-compose up --build
```

3. Access the application:
Open your browser and navigate to `http://localhost:8080`

## Development

To run the application locally without Docker:

1. Install dependencies:
```bash
go mod download
```

2. Run the application:
```bash
go run cmd/main.go
```

## Testing

Run the tests:
```bash
go test ./...
```

## License

[License details]
