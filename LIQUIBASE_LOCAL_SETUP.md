# Liquibase Local Setup Guide

This guide will help you set up and run Liquibase migrations locally instead of using Docker.

## Prerequisites

### 1. Install Liquibase

**On macOS:**
```bash
brew install liquibase
```

**On Linux/Windows:**
Download from [Liquibase Releases](https://github.com/liquibase/liquibase/releases) and add to your PATH.

### 2. Install PostgreSQL JDBC Driver

Liquibase needs the PostgreSQL JDBC driver to connect to PostgreSQL.

**Option A: Download manually**
```bash
# Create lib directory for Liquibase
mkdir -p ~/.liquibase/lib

# Download PostgreSQL JDBC driver
wget https://jdbc.postgresql.org/download/postgresql-42.6.0.jar -O ~/.liquibase/lib/postgresql-42.6.0.jar
```

**Option B: Use Homebrew (macOS)**
```bash
# The JDBC driver is usually included with Liquibase installation via Homebrew
```

### 3. Start PostgreSQL

**Using Docker Compose (recommended):**
```bash
# Start only PostgreSQL from the docker-compose
docker-compose up postgres -d
```

**Or install PostgreSQL locally and make sure it's running on port 5433**

## Running Liquibase Migrations

### Method 1: Using the Script (Recommended)

```bash
# Run the automated script
./scripts/run-liquibase.sh
```

### Method 2: Manual Execution

```bash
# 1. Navigate to project root
cd /path/to/TakeHomeAssignment

# 2. Create database if it doesn't exist
createdb -h localhost -p 5433 -U postgres internal_transfers

# 3. Run Liquibase migrations
liquibase --defaults-file=db/liquibase/liquibase.properties update
```

### Method 3: Direct Command with Parameters

```bash
liquibase \
  --driver=org.postgresql.Driver \
  --url=jdbc:postgresql://localhost:5433/internal_transfers \
  --username=postgres \
  --password=password \
  --changelog-file=db/liquibase/changelog.xml \
  update
```

## Verification

After running the migrations, verify the tables were created:

```bash
# Connect to the database
psql -h localhost -p 5433 -U postgres -d internal_transfers

# List tables
\dt

# Should show:
#  public | accounts     | table | postgres
#  public | transactions | table | postgres

# Check table structure
\d accounts
\d transactions

# Exit
\q
```

## Configuration Files

### Current Liquibase Configuration

**File: `db/liquibase/liquibase.properties`**
```properties
# Liquibase Configuration for Local Execution
driver=org.postgresql.Driver
url=jdbc:postgresql://localhost:5433/internal_transfers
username=postgres
password=password
changeLogFile=db/liquibase/changelog.xml
logLevel=INFO
classpath=db/liquibase
```

### Application Configuration

**File: `.env` (create from .env.example)**
```bash
# Database Configuration  
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=internal_transfers
DB_SSLMODE=disable

# Server Configuration
PORT=8080
GIN_MODE=debug
```

## Troubleshooting

### Issue: "ClassNotFoundException: org.postgresql.Driver"

**Solution:** Install PostgreSQL JDBC driver
```bash
# Download the driver
curl -o ~/.liquibase/lib/postgresql-42.6.0.jar \
  https://jdbc.postgresql.org/download/postgresql-42.6.0.jar

# Or add to LIQUIBASE_CLASSPATH
export LIQUIBASE_CLASSPATH=~/.liquibase/lib/postgresql-42.6.0.jar
```

### Issue: "Connection refused"

**Solution:** Make sure PostgreSQL is running
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5433 -U postgres

# If not running, start it
docker-compose up postgres -d
```

### Issue: "Database does not exist"

**Solution:** Create the database first
```bash
createdb -h localhost -p 5433 -U postgres internal_transfers
```

### Issue: "File not found" errors

**Solution:** Run from project root directory
```bash
cd /Users/urvishsojitra/Documents/TakeHomeAssignment
liquibase --defaults-file=db/liquibase/liquibase.properties update
```

## Running the Application

After successful Liquibase migration:

```bash
# 1. Make sure you have .env file configured
cp .env.example .env
# Edit .env with correct database settings

# 2. Run the application
go run cmd/server/main.go

# 3. Test the setup
curl http://localhost:8080/health
# Expected: OK
```

## Clean Up (Optional)

To reset and re-run migrations:

```bash
# Drop and recreate database
dropdb -h localhost -p 5433 -U postgres internal_transfers
createdb -h localhost -p 5433 -U postgres internal_transfers

# Re-run migrations
liquibase --defaults-file=db/liquibase/liquibase.properties update
```

## Quick Commands Reference

```bash
# Check Liquibase status
liquibase --defaults-file=db/liquibase/liquibase.properties status

# Validate changelog
liquibase --defaults-file=db/liquibase/liquibase.properties validate

# Show SQL that would be executed
liquibase --defaults-file=db/liquibase/liquibase.properties update-sql

# Rollback last changeset
liquibase --defaults-file=db/liquibase/liquibase.properties rollback-count 1
```