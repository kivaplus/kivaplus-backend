# ⚡ Quick Start - Database Connection

**For experienced developers who just need the commands:**

## 🚀 One-Time Setup

```bash
# 1. Deploy infrastructure (if needed)
cd infra && cdk deploy

# 2. Start SSH tunnel (keep running in separate terminal)
cd kivaplus-backend && ./scripts/connect-db.sh

# 3. Load credentials (run in each new terminal)
source ./scripts/get-db-credentials.sh

# 4. Run migrations
./scripts/run-migrations.sh up

# 5. Test connection
make test-postgres
```

## 📋 Daily Commands

```bash
# Connect to database
source ./scripts/get-db-credentials.sh

# Check migration status
./scripts/run-migrations.sh version

# Apply new migrations
./scripts/run-migrations.sh up

# Test database
make test-postgres

# Check schema
./scripts/check-db-schema.sh
```

## 🔧 Troubleshooting

```bash
# Test connection
./scripts/test-db-connection.sh

# Debug issues
./scripts/debug-rds.sh

# Alternative connection
make port-forward
```

**Need more details?** See [HELPME.md](HELPME.md) for the complete guide.
