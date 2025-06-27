# Quick Liquibase Local Commands

## 🚀 Quick Setup

1. **Install Liquibase:**
   ```bash
   brew install liquibase
   ```

2. **Start PostgreSQL:**
   ```bash
   docker-compose up postgres -d
   ```

3. **Run Migrations:**
   ```bash
   # From project root directory
   liquibase --defaults-file=db/liquibase/liquibase.properties update
   ```

## ✅ Step-by-Step Execution

```bash
# 1. Navigate to project directory
cd /Users/urvishsojitra/Documents/TakeHomeAssignment

# 2. Start PostgreSQL (if not already running)
docker-compose up postgres -d

# 3. Wait for PostgreSQL to be ready
docker-compose logs postgres

# 4. Create database (if not exists)
createdb -h localhost -p 5433 -U postgres internal_transfers

# 5. Run Liquibase migrations
liquibase --defaults-file=db/liquibase/liquibase.properties update

# 6. Verify tables were created
psql -h localhost -p 5433 -U postgres -d internal_transfers -c "\dt"

# 7. Run the application
go run cmd/server/main.go
```

## 🔧 Alternative Direct Command

If the properties file doesn't work, use direct parameters:

```bash
liquibase \
  --driver=org.postgresql.Driver \
  --url=jdbc:postgresql://localhost:5433/internal_transfers \
  --username=postgres \
  --password=password \
  --changelog-file=db/liquibase/changelog.xml \
  update
```

## 📊 Verify Success

```sql
-- Connect to database
psql -h localhost -p 5433 -U postgres -d internal_transfers

-- Check tables
\dt

-- Expected output:
--           List of relations
--  Schema |    Name      | Type  |  Owner   
-- --------+--------------+-------+----------
--  public | accounts     | table | postgres
--  public | transactions | table | postgres

-- Check accounts table structure
\d accounts

-- Check transactions table structure  
\d transactions

-- Exit
\q
```