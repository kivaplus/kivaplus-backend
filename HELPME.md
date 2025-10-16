# 🚀 KivaPlus Backend - Database Setup Guide

This guide will help you connect to the database and run migrations from scratch. Follow these steps if you're setting up the project for the first time.

## 📋 Prerequisites

Before starting, make sure you have:

- [x] **AWS CLI** installed and configured
- [x] **Go 1.24+** installed
- [x] **PostgreSQL client** (psql) installed
- [x] **SSH client** available
- [x] **AWS CDK** deployed (infrastructure must be running)

## 🔧 Step 1: Verify AWS Configuration

First, ensure you're authenticated with AWS and can access the resources:

```bash
# Check AWS authentication
aws sts get-caller-identity

# Verify you can access the region (default: us-east-1)
aws ec2 describe-instances --region us-east-1 --query 'Reservations[].Instances[].{Name:Tags[?Key==`Name`].Value|[0],State:State.Name}' --output table
```

**Expected output**: You should see your AWS account info and any running EC2 instances.

## 🏗️ Step 2: Deploy Infrastructure (if not done)

If the infrastructure isn't deployed yet:

```bash
cd infra
cdk deploy --require-approval never
```

This creates:
- 🌐 VPC with public/private subnets
- 🐘 RDS PostgreSQL database
- 🖥️ Bastion host for secure access
- 🔐 Secrets Manager for database credentials

## 🔑 Step 3: Set Up Database Connection

### Option A: Session Manager (Recommended)

**Best for development - no SSH keys, fully secure, works from anywhere!**

1. **Install Session Manager plugin** (one-time setup):
```bash
# macOS with Homebrew
brew install --cask session-manager-plugin

# Or download from AWS documentation
```

2. **Connect to database instance**:
```bash
cd kivaplus-backend
make connect-ssm
```

This opens a secure shell session where you can:
- Connect directly to PostgreSQL: `~/scripts/connect-db.sh`
- Set up environment: `source ~/scripts/get-db-env.sh`
- Run migrations: `go run ./cmd/migrate -action=up`

### Option B: SSH Tunnel (Legacy)

If you prefer the old SSH method:

```bash
cd kivaplus-backend
./scripts/connect-db.sh
source ./scripts/get-db-credentials.sh
```

### Option C: Port Forwarding

For local development with your own tools:

```bash
cd kivaplus-backend
make port-forward
source ./scripts/get-db-credentials.sh
```

## 🧪 Step 4: Test Database Connection

Verify everything is working:

```bash
cd kivaplus-backend
./scripts/test-db-connection.sh
```

**Expected output**:
```
✅ Túnel SSH ativo na porta 5433
✅ Conexão com psql bem-sucedida!
✅ Conexão com migrate bem-sucedida!
```

If you see errors, check the [Troubleshooting](#-troubleshooting) section below.

## 📊 Step 5: Run Database Migrations

Now apply all database migrations to create the schema:

```bash
cd kivaplus-backend

# Check current migration status
./scripts/run-migrations.sh version

# Apply all migrations
./scripts/run-migrations.sh up
```

**Expected output**:
```
✅ Applied 000001_create_function.up.sql
✅ Applied 000002_create_endereco.up.sql
✅ Applied 000003_create_pessoa.up.sql
... (17 migrations total)
```

## ✅ Step 6: Verify Database Schema

Check that all tables were created:

```bash
cd kivaplus-backend
./scripts/check-db-schema.sh
```

You should see 17 tables including:
- `usuario` (users)
- `pessoa` (people)
- `endereco` (addresses)
- `condominio` (condominiums)
- And more...

## 🧪 Step 7: Test the Database Client

Run the PostgreSQL client tests to ensure everything works:

```bash
cd kivaplus-backend
make test-postgres
```

**Expected output**:
```
✅ PostgreSQL client created successfully
✅ User test-user@kivaplus.com inserted successfully
✅ Password validation passed
... All tests should pass
```

## 🎯 Step 8: You're Ready!

Congratulations! Your database is now set up and ready. You can:

- ✅ Connect to the database
- ✅ Run migrations
- ✅ Use the PostgreSQL client in your code
- ✅ Run tests

## 📚 Common Commands Reference

### Database Connection
```bash
# Start SSH tunnel (keep running)
./scripts/connect-db.sh

# Load credentials (run in each new terminal)
source ./scripts/get-db-credentials.sh

# Test connection
./scripts/test-db-connection.sh
```

### Migrations
```bash
# Check current version
./scripts/run-migrations.sh version

# Apply migrations
./scripts/run-migrations.sh up

# Rollback migrations (careful!)
./scripts/run-migrations.sh down 1

# Check database schema
./scripts/check-db-schema.sh
```

### Testing
```bash
# Test PostgreSQL client
make test-postgres

# Test specific functionality
make test-postgres-specific

# Manual database test
go run ./cmd/test-postgres/main.go
```

### Using Makefile
```bash
# See all available commands
make help

# Quick migration commands
make migrate-local-version
make migrate-local-up
```

## 🔧 Troubleshooting

### ❌ "SSH connection failed"

**Problem**: Can't connect to bastion host
```bash
# Check if bastion host exists
aws ec2 describe-instances --region us-east-1 --filters "Name=tag:Name,Values=*Bastion*"

# Try Systems Manager instead
make port-forward
```

### ❌ "DATABASE_URL not set"

**Problem**: Environment variables not loaded
```bash
# Always run this in each new terminal
source ./scripts/get-db-credentials.sh

# Verify it's set
echo $DATABASE_URL
```

### ❌ "Connection refused on port 5433"

**Problem**: SSH tunnel not running
```bash
# Check if tunnel is active
lsof -i :5433

# Restart tunnel
./scripts/connect-db.sh
```

### ❌ "pq: no pg_hba.conf entry"

**Problem**: SSL/connection configuration
```bash
# Check database name (should be 'kivaplusdb', not 'kivaplus_db')
echo $DATABASE_URL

# Try updating RDS configuration
cd infra && cdk deploy
```

### ❌ "Migration failed"

**Problem**: Database schema conflicts
```bash
# Check current schema
./scripts/check-db-schema.sh

# If needed, drop conflicting tables
psql "$DATABASE_URL" -c "DROP TABLE IF EXISTS schema_migrations CASCADE;"

# Re-run migrations
./scripts/run-migrations.sh up
```

### ❌ "AWS credentials not found"

**Problem**: AWS CLI not configured
```bash
# Configure AWS CLI
aws configure

# Or use environment variables
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
export AWS_DEFAULT_REGION=us-east-1
```

## 🆘 Getting Help

If you're still having issues:

1. **Check the logs**: Most scripts provide detailed error messages
2. **Verify infrastructure**: Ensure `cdk deploy` completed successfully
3. **Check AWS console**: Verify RDS instance is running and accessible
4. **Test step by step**: Use the individual test scripts to isolate issues

### Debug Commands
```bash
# Debug RDS endpoints
./scripts/debug-rds.sh

# Debug security groups
./scripts/debug-security-groups.sh

# Test SSH tunnel specifically
./scripts/test-tunnel.sh
```

## 📖 Next Steps

Once your database is set up:

1. **Explore the schema**: Check `internal/database/migrations/` for table definitions
2. **Use the client**: See `lambda/database/README.md` for usage examples
3. **Add new migrations**: Use `./create_migrate.sh` to add new schema changes
4. **Deploy to Lambda**: The database client automatically works in AWS Lambda

## 🔗 Related Documentation

- [Database Client README](lambda/database/README.md) - How to use the PostgreSQL client
- [Migration System](internal/database/migrations.go) - How migrations work
- [Infrastructure Code](../infra/) - AWS CDK infrastructure setup

---

**Happy coding!** 🎉 If you encounter any issues not covered here, please update this guide to help future developers.
