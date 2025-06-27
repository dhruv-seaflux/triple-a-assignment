#!/bin/bash
set -e

echo "🔧 Fixing Liquibase setup and running migrations..."

# Go to project root
cd "$(dirname "$0")"

echo "📍 Current directory: $(pwd)"

# Start PostgreSQL if not running
echo "🐘 Starting PostgreSQL..."
docker-compose up postgres -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Check if PostgreSQL is ready
if ! pg_isready -h localhost -p 5433 -U postgres &> /dev/null; then
    echo "❌ PostgreSQL is not ready. Please check Docker logs:"
    echo "   docker-compose logs postgres"
    exit 1
fi

# Create database if not exists
echo "📝 Creating database if it doesn't exist..."
createdb -h localhost -p 5433 -U postgres internal_transfers 2>/dev/null || echo "Database already exists"

# Download JDBC driver if not exists
echo "📥 Checking PostgreSQL JDBC driver..."
mkdir -p ~/.liquibase/lib
if [ ! -f ~/.liquibase/lib/postgresql-42.6.0.jar ]; then
    echo "📥 Downloading PostgreSQL JDBC driver..."
    curl -L -o ~/.liquibase/lib/postgresql-42.6.0.jar \
      https://jdbc.postgresql.org/download/postgresql-42.6.0.jar
fi

# Set classpath environment variable
export LIQUIBASE_CLASSPATH=~/.liquibase/lib/postgresql-42.6.0.jar

echo "🚀 Running Liquibase migrations with absolute paths..."

# Run with absolute paths to avoid file not found issues
liquibase \
  --driver=org.postgresql.Driver \
  --url=jdbc:postgresql://localhost:5433/internal_transfers \
  --username=postgres \
  --password=password \
  --changelog-file="$(pwd)/db/liquibase/changelog.xml" \
  --log-level=INFO \
  update

echo ""
echo "✅ Liquibase migrations completed successfully!"
echo ""
echo "🔍 Verifying tables were created..."
psql -h localhost -p 5433 -U postgres -d internal_transfers -c "\dt"

echo ""
echo "🎉 Setup complete! You can now run the application:"
echo "   go run cmd/server/main.go"