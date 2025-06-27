# Liquibase Troubleshooting Guide

## Common Error Solutions

### 1. **JDBC Driver Not Found Error**

**Error:**
```
ClassNotFoundException: org.postgresql.Driver
```

**Solutions:**

**Option A: Download JDBC Driver**
```bash
# Create lib directory
mkdir -p ~/.liquibase/lib

# Download PostgreSQL JDBC driver
curl -o ~/.liquibase/lib/postgresql-42.6.0.jar \
  https://jdbc.postgresql.org/download/postgresql-42.6.0.jar

# Set classpath
export LIQUIBASE_CLASSPATH=~/.liquibase/lib/postgresql-42.6.0.jar
```

**Option B: Use with Liquibase Hub**
```bash
# Install with driver included
brew reinstall liquibase
```

### 2. **Connection Refused Error**

**Error:**
```
Connection refused: connect
```

**Solutions:**
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5433 -U postgres

# If not running, start PostgreSQL
docker-compose up postgres -d

# Wait for it to be ready
docker-compose logs postgres | grep "ready to accept connections"
```

### 3. **Database Does Not Exist Error**

**Error:**
```
database "internal_transfers" does not exist
```

**Solution:**
```bash
# Create the database
createdb -h localhost -p 5433 -U postgres internal_transfers

# Or connect and create manually
psql -h localhost -p 5433 -U postgres -c "CREATE DATABASE internal_transfers;"
```

### 4. **File Not Found Error**

**Error:**
```
FileNotFoundException: changelogs/001-create-accounts-table.xml
```

**Solution:**
```bash
# Make sure you're in the project root directory
cd /Users/urvishsojitra/Documents/TakeHomeAssignment

# Check if files exist
ls -la db/liquibase/changelogs/

# Run with absolute path
liquibase --defaults-file=$(pwd)/db/liquibase/liquibase.properties update
```

### 5. **Authentication Failed Error**

**Error:**
```
password authentication failed for user "postgres"
```

**Solution:**
```bash
# Check Docker container logs
docker-compose logs postgres

# Try connecting manually to verify credentials
psql -h localhost -p 5433 -U postgres -d postgres

# If password is different, update liquibase.properties
```

## Step-by-Step Debugging

### 1. **Check PostgreSQL Connection**

```bash
# Test basic connection
psql -h localhost -p 5433 -U postgres -d postgres -c "SELECT version();"

# If this fails, PostgreSQL is not accessible
```

### 2. **Check Database Exists**

```bash
# List databases
psql -h localhost -p 5433 -U postgres -l | grep internal_transfers

# Create if missing
createdb -h localhost -p 5433 -U postgres internal_transfers
```

### 3. **Test Liquibase Installation**

```bash
# Check Liquibase version
liquibase --version

# Should show something like: Liquibase Version: 4.x.x
```

### 4. **Validate Changelog**

```bash
# Validate syntax without connecting to database
liquibase --defaults-file=db/liquibase/liquibase.properties validate
```

### 5. **Check File Paths**

```bash
# Verify all files exist
ls -la db/liquibase/changelog.xml
ls -la db/liquibase/changelogs/
```

## Alternative Working Commands

### Method 1: With Full Paths

```bash
cd /Users/urvishsojitra/Documents/TakeHomeAssignment

liquibase \
  --driver=org.postgresql.Driver \
  --url=jdbc:postgresql://localhost:5433/internal_transfers \
  --username=postgres \
  --password=password \
  --changelog-file=$(pwd)/db/liquibase/changelog.xml \
  --log-level=INFO \
  update
```

### Method 2: Using Environment Variables

```bash
export LIQUIBASE_COMMAND_URL=jdbc:postgresql://localhost:5433/internal_transfers
export LIQUIBASE_COMMAND_USERNAME=postgres
export LIQUIBASE_COMMAND_PASSWORD=password
export LIQUIBASE_COMMAND_CHANGELOG_FILE=db/liquibase/changelog.xml

liquibase update
```

### Method 3: Fix Properties File Paths

Update `db/liquibase/liquibase.properties`:
```properties
# Use absolute paths
driver=org.postgresql.Driver
url=jdbc:postgresql://localhost:5433/internal_transfers
username=postgres
password=password
changeLogFile=/Users/urvishsojitra/Documents/TakeHomeAssignment/db/liquibase/changelog.xml
logLevel=INFO
```

## Quick Diagnostic Commands

```bash
# 1. Check PostgreSQL is running
docker ps | grep postgres

# 2. Check PostgreSQL port
netstat -an | grep 5433

# 3. Test database connection
pg_isready -h localhost -p 5433 -U postgres

# 4. Check Liquibase can find files
liquibase --defaults-file=db/liquibase/liquibase.properties status

# 5. See what Liquibase would execute
liquibase --defaults-file=db/liquibase/liquibase.properties update-sql
```

## Emergency Fix Script

Create a file `fix-liquibase.sh`:

```bash
#!/bin/bash
set -e

echo "🔧 Fixing Liquibase setup..."

# Go to project root
cd /Users/urvishsojitra/Documents/TakeHomeAssignment

# Start PostgreSQL if not running
docker-compose up postgres -d

# Wait for PostgreSQL
sleep 5

# Create database if not exists
createdb -h localhost -p 5433 -U postgres internal_transfers 2>/dev/null || true

# Download JDBC driver if not exists
mkdir -p ~/.liquibase/lib
if [ ! -f ~/.liquibase/lib/postgresql-42.6.0.jar ]; then
    echo "📥 Downloading PostgreSQL JDBC driver..."
    curl -o ~/.liquibase/lib/postgresql-42.6.0.jar \
      https://jdbc.postgresql.org/download/postgresql-42.6.0.jar
fi

# Set classpath
export LIQUIBASE_CLASSPATH=~/.liquibase/lib/postgresql-42.6.0.jar

# Run with absolute paths
liquibase \
  --driver=org.postgresql.Driver \
  --url=jdbc:postgresql://localhost:5433/internal_transfers \
  --username=postgres \
  --password=password \
  --changelog-file=$(pwd)/db/liquibase/changelog.xml \
  update

echo "✅ Liquibase migrations completed!"
```

## Run the Fix

```bash
chmod +x fix-liquibase.sh
./fix-liquibase.sh
```