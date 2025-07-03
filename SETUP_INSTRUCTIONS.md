# GM58-HR Backend - Setup Instructions

## Complete Setup Guide

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 12+
- Redis 6+
- Git

### Quick Setup

1. **Navigate to project directory**
   ```bash
   cd gm58-hr-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

4. **Database setup**
   ```sql
   CREATE DATABASE gm58_hr;
   CREATE USER hr_user WITH PASSWORD 'hr_password';
   GRANT ALL PRIVILEGES ON DATABASE gm58_hr TO hr_user;
   ```

5. **Run migrations**
   ```bash
   go run scripts/migrate.go -cmd=up
   ```

6. **Start the server**
   ```bash
   go run cmd/server/main.go
   # OR
   make run
   ```

### Using Docker (Recommended)

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f backend
```

### Testing

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Login (default admin)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Default Data

After setup, you'll have:
- Admin user: admin/admin123
- Currencies: USD, ZWL, ZAR, GBP, EUR
- Leave types: Annual, Sick, Maternity, Paternity, etc.
- Default departments: HR, Finance, IT, Sales, Marketing, Operations

The backend will be available at http://localhost:8080
