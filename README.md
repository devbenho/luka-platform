# Luka Platform

A modern e-commerce platform built with Go and Next.js, featuring a robust backend API and a responsive frontend interface.

## 🚀 Features

### Backend (Go)
- RESTful API built with Gin framework
- MongoDB database integration
- JWT-based authentication
- Modular architecture with clean separation of concerns
- Error handling and validation
- Docker support for containerization

### Frontend (Next.js)
- Modern React-based UI
- Tailwind CSS for styling
- TypeScript for type safety
- Responsive design

### Core Modules
- User Management
- Product Catalog
- Inventory Management
- Order Processing
- Warehouse Management
- Category Management
- Store Management

## 🛠️ Technical Stack

### Backend
- Go 1.23
- Gin Web Framework
- MongoDB
- JWT Authentication
- Docker

### Frontend
- Next.js
- TypeScript
- Tailwind CSS
- React

## 📦 Getting Started

### Prerequisites
- Go 1.23 or later
- Node.js 18+ and npm (frontend only)
- MongoDB 4.4+ running as a **replica set**: orders use multi-document transactions. A single-node replica set is fine for development. The API refuses to start against a standalone `mongod`.
- Docker (optional; the easiest way to get MongoDB)

### Get the code

```bash
git clone https://github.com/devbenho/luka-platform.git
cd luka-platform
```

### Configuration

The API reads these environment variables. It also loads a `.env` file from the working directory if one exists; variables already set in the environment take precedence.

| Variable | Required | Description |
|---|---|---|
| `PORT` | yes | HTTP port, e.g. `2707` |
| `MONGO_URI` | yes | MongoDB connection string, e.g. `mongodb://localhost:27017/?directConnection=true` |
| `DB_NAME` | yes | Database name, e.g. `luka` |
| `JWT_SECRET` | yes | Secret used to sign access tokens. Use a long random value, e.g. `openssl rand -hex 32`. |
| `JWT_TYPE` | no | Token type label |
| `ALLOWED_ORIGINS` | no | CORS origins |

If a required variable is missing, the API exits at startup.

### Run with Docker Compose

This starts MongoDB as a single-node replica set, then the API on port 2707:

```bash
export JWT_SECRET=$(openssl rand -hex 32)
docker compose up
```

- Swagger UI: http://localhost:2707/swagger/index.html
- Health check: `curl localhost:2707/ping`

### Run locally

1. Start MongoDB as a single-node replica set:

```bash
docker run -d --name luka-mongo -p 27017:27017 mongo:7 --replSet rs0 --bind_ip_all
```

```bash
docker exec luka-mongo mongosh --quiet --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'localhost:27017'}]})"
```

2. Configure and run the API:

```bash
cp .env.example .env   # then set the variables from the table above
go run ./cmd/api
```

### Tests

Unit tests need no database:

```bash
go test ./...
```

Integration tests drive the real HTTP routes against MongoDB, each run in a throwaway database. They need the replica set from "Run locally":

```bash
LUKA_TEST_MONGO_URI="mongodb://localhost:27017/?directConnection=true" go test -tags integration ./...
```

CI (`.github/workflows/go.yml`) runs `go vet`, the unit tests with `-race`, and the integration tests against a MongoDB replica set on every pull request.

### API documentation

The Swagger spec in `docs/` is generated from the handler annotations. Regenerate it after changing handlers or DTOs:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g cmd/api/main.go -o docs
```

### Frontend

```bash
cd web
npm install
npm run dev
```

## 🔍 Project Structure

```
.
├── cmd/              # Application entry points
├── internal/         # Core application code
│   ├── user/        # User management
│   ├── product/     # Product management
│   ├── orders/      # Order processing
│   ├── inventory/   # Inventory management
│   ├── warehouse/   # Warehouse operations
│   ├── category/    # Category management
│   ├── store/       # Store management
│   └── server/      # Server configuration
├── web/             # Frontend application
├── pkg/             # Shared packages
└── docs/            # Documentation
```

## 🎯 Improvements & Future Features

### Backend Improvements
1. Implement comprehensive API documentation using Swagger/OpenAPI
2. Add rate limiting and request validation
3. Implement caching layer (Redis)
4. Add comprehensive logging and monitoring
5. Implement background job processing
6. Add more robust error handling and recovery
7. Implement database migrations

### Frontend Improvements
1. Implement state management (Redux/Context)
2. Add comprehensive form validation
3. Implement proper error boundaries
4. Add loading states and skeletons
5. Implement proper authentication flow
6. Add comprehensive testing
7. Implement proper SEO optimization
8. Add PWA support

### General Improvements
1. Implement CI/CD pipeline
2. Add comprehensive monitoring
3. Implement proper backup strategy
4. Add comprehensive security measures
5. Implement proper documentation
6. Add performance optimization
7. Implement proper logging
8. Add comprehensive analytics

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- Dev Benho - Initial work

## 🙏 Acknowledgments

- Thanks to all contributors who have helped shape this project 