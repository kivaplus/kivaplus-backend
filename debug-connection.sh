#!/bin/bash

# Debug Database Connection
echo "🔍 Database Connection Debug"
echo "==========================="

# Check if DATABASE_URL is set
echo "📋 Environment Check:"
echo "DATABASE_URL: $DATABASE_URL"

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL is not set!"
    echo "💡 Setting default DATABASE_URL..."
    export DATABASE_URL='postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable'
    echo "✅ DATABASE_URL set to: $DATABASE_URL"
fi

echo ""
echo "🐳 Docker Container Check:"
docker compose ps postgres

echo ""
echo "🧪 Direct PostgreSQL Connection Test:"
if docker compose exec postgres psql -U kivaplus -d kivaplus_local -c "SELECT 'Direct connection works' as test;" 2>/dev/null; then
    echo "✅ Direct PostgreSQL connection works"
else
    echo "❌ Direct PostgreSQL connection failed"
    exit 1
fi

echo ""
echo "🔍 Go Database Connection Debug:"
go run -tags integration ./cmd/debug-db/main.go

echo ""
echo "🧪 Original Check (for comparison):"
go run -tags integration ./cmd/check-db/main.go
