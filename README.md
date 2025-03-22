# E-Commerce Backend Microservices

A modern, scalable e-commerce backend built with Go microservices architecture. This project follows Domain-Driven Design (DDD) and Clean Architecture principles to create a maintainable and extensible e-commerce platform.

## Architecture Overview

This application is structured as three primary microservices:

- **User Service**: Authentication, user management, and profiles
- **Product Service**: Product catalog, categories, and inventory management
- **Order Service**: Order processing, payments, shipping, and notifications

Each service is built with a clean architecture approach:
- Domain Layer: Core business entities and rules
- Application Layer: Orchestrates business logic (usecases)
- Adapter Layer: Controllers and repositories
- Infrastructure Layer: Database, messaging, and external services

## Technologies

- **Language**: Go (version 1.23)
- **API**: Dual API support with Fiber (REST) and gRPC
- **Databases**: MySQL (via GORM), Cassandra (via gocql)
- **Messaging**: Kafka for event-driven architecture
- **API Gateway**: Kong
- **Authentication**: JWT-based
- **Logging**: Zap and Logrus
- **Containerization**: Docker
- **Orchestration**: Kubernetes

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose
- Protobuf compiler (for gRPC)
- Make

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/hydr0g3nz/ecom_back_microservice.git
   cd ecom_back_microservice
   ```

2. Install required tools:
   ```
   make install-tools
   ```

3. Generate gRPC code:
   ```
   make proto-gen
   ```

4. Start the development environment:
   ```
   cd deployments/docker
   docker-compose up -d
   ```

5. Build the services:
   ```
   make build
   ```

### Running the Services

You can run each service individually:

```
make run-user     # Run user service
make run-product  # Run product service
make run-order    # Run order service
```

Or run all services:

```
make run
```

### Testing

Run tests for specific services:

```
make test-user
make test-product
make test-order
```

Or run all tests:

```
make test
```

## Service Details

### User Service

The User Service handles authentication, authorization, and user management:

- User registration and login
- JWT token generation and validation
- User profile management
- Role-based access control

### Product Service

The Product Service manages the product catalog:

- Product CRUD operations
- Category management
- Inventory tracking
- Product search and filtering

### Order Service

The Order Service manages the complete order lifecycle:

- Order creation and processing
- Payment handling
- Shipping management
- Order status updates and notifications
- Event-driven updates via Kafka

## Deployment

### Docker

A Docker Compose configuration is provided for local development in the `deployments/docker` directory.

### Kubernetes

Kubernetes manifests are available in the `deployments/k8s` directory for production deployment.

## Project Structure

```
├── cmd/                  # Service entry points
│   ├── user_service/
│   ├── product_service/
│   └── order_service/
├── internal/             # Private application code
│   ├── user_service/
│   │   ├── adapter/      # Controllers and repositories
│   │   ├── config/       # Service configuration
│   │   ├── domain/       # Domain models and logic
│   │   └── usecase/      # Business logic
│   ├── product_service/
│   └── order_service/
├── pkg/                  # Shared utilities
├── deployments/          # Deployment configurations
│   ├── docker/
│   └── k8s/
├── docs/                 # Documentation
├── scripts/              # Build and automation scripts
└── tests/                # Integration and e2e tests
```

## API Documentation

API documentation for both REST and gRPC interfaces is available in the `docs` directory.

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/my-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
