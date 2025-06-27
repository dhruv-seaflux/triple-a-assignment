#!/bin/bash

# Script to run Liquibase migrations locally
# Make sure you have Liquibase installed and PostgreSQL running

set -e

echo "🔧 Running Liquibase migrations locally..."

# Change to the project root directory
cd "$(dirname "$0")/.."

# Check if Liquibase is installed
if ! command -v liquibase &> /dev/null; then
    echo "❌ Liquibase is not installed. Please install it first:"
    echo ""
    echo "On macOS with Homebrew:"
    echo "  brew install liquibase"
    echo ""
    echo "On other systems, download from:"
    echo "  https://github.com/liquibase/liquibase/releases"
    echo ""
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p 5433 -U postgres &> /dev/null; then
    echo "❌ PostgreSQL is not running on localhost:5433"
    echo ""
    echo "Please start PostgreSQL first:"
    echo "  docker-compose up postgres -d"
    echo ""
    echo "Or if you have PostgreSQL installed locally:"
    echo "  Make sure it's running on port 5433"
    echo ""
    exit 1
fi

# Check if database exists
if ! psql -h localhost -p 5433 -U postgres -lqt | cut -d \| -f 1 | grep -qw internal_transfers; then
    echo "📝 Creating database 'internal_transfers'..."
    createdb -h localhost -p 5433 -U postgres internal_transfers
fi

# Run Liquibase update
echo "🚀 Running Liquibase update..."
echo "Configuration file: db/liquibase/liquibase.properties"
echo "Changelog file: db/liquibase/changelog.xml"
echo ""

liquibase --defaults-file=db/liquibase/liquibase.properties update

echo ""
echo "✅ Liquibase migrations completed successfully!"
echo ""
echo "You can now run the application:"
echo "  go run cmd/server/main.go"