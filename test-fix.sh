#!/bin/bash

# Test the package fix
export DATABASE_URL='postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable'

echo ""
echo "🧪 Testing integration tests compilation..."
go test -tags integration ./tests/integration -run=NonExistentTest -v

echo ""
echo "✅ Package conflict fixed!"
