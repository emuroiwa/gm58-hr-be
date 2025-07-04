# GM58-HR Payroll Management System - Backend

A comprehensive payroll management system built with Go, supporting multi-currency operations and tax compliance.

## Features

- **Multi-currency Support**: Handle payroll in USD, ZWL, ZAR, GBP, EUR
- **Tax Compliance**: PAYE, AIDS Levy, NSSA calculations
- **Employee Management**: Complete employee lifecycle management
- **Payroll Processing**: Automated payroll calculations and processing
- **Leave Management**: Track and manage employee leave requests
- **Audit Logging**: Complete audit trail for all operations
- **RESTful API**: Well-documented REST API for frontend integration
- **Authentication**: JWT-based authentication with role-based access

## Technology Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis for caching and job queues
- **Authentication**: JWT tokens
- **Currency**: Real-time exchange rate integration

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12+
- Redis 6+
- Git

### Development Setup

1. **Clone and setup**
   ```bash
   # Navigate to project directory
   cd gm58-hr-backend
   
   # Install dependencies
   go mod download
   
   # Set up environment
   cp .env.example .env
   # Edit .env with your configuration
   
   # Run development setup
   ./scripts/setup-dev.sh
   
   # Start the server
   go run cmd/server/main.go
   ```

2. **Using Docker**
   ```bash
   # Start all services
   docker-compose up -d
   
   # Check status
   docker-compose ps
   
   # View logs
   docker-compose logs -f backend
   ```

### API Documentation

#### Authentication
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","email":"user@example.com","password":"password123"}'
```

#### Employee Management
```bash
# Get employees
curl http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer YOUR_TOKEN"

# Create employee
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@company.com",
    "position_id": 1,
    "department_id": 1,
    "basic_salary": 5000.00,
    "currency_id": 1,
    "hire_date": "2024-01-01T00:00:00Z"
  }'
```

#### Payroll Operations
```bash
# Create payroll period
curl -X POST http://localhost:8080/api/v1/payroll/periods \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"year": 2024, "month": 12, "description": "December 2024 Payroll"}'

# Process payroll
curl -X POST http://localhost:8080/api/v1/payroll/periods/1/process \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Currency Operations
```bash
# Get exchange rate
curl "http://localhost:8080/api/v1/currencies/exchange-rate?from=USD&to=ZAR" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Convert amount
curl "http://localhost:8080/api/v1/currencies/convert?amount=1000&from=USD&to=ZAR" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Default Credentials

- Username: `admin`
- Password: `admin123`

### Development Commands

```bash
make build           # Build the application
make run             # Run the application
make test            # Run tests
make docker-up       # Start Docker containers
make migrate-up      # Run database migrations
make setup-dev       # Setup development environment
```

### Project Structure

```
gm58-hr-backend/
├── cmd/server/              # Application entry point
├── internal/
│   ├── api/handlers/        # HTTP request handlers
│   ├── config/             # Configuration management
│   ├── database/           # Database connection
│   ├── models/             # Data models
│   ├── services/           # Business logic
│   └── middleware/         # Authentication, CORS
├── pkg/                    # Reusable packages
├── migrations/             # Database migrations
├── tests/                  # Test files
└── scripts/               # Build and deployment scripts
```

## Multi-Currency Support

The system supports multiple currencies with real-time exchange rates:

- Base currency: USD
- Supported currencies: USD, ZWL, ZAR, GBP, EUR
- Automatic exchange rate updates
- Currency conversion for reporting
- Employee salary in preferred currency

## Tax Calculations

### PAYE (Pay As You Earn)
Based on tax brackets:
- $0 - $100: 0%
- $100.01 - $300: 20%
- $300.01 - $1,000: 25%
- $1,000.01 - $2,000: 30%
- $2,000.01 - $3,000: 35%
- $3,000.01+: 40%

### AIDS Levy
3% of PAYE tax

### NSSA Contribution
3% of gross salary (employee contribution)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

This project is licensed under the MIT License.
# gm58-hr-be
