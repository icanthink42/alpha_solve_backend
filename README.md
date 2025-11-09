# Alpha Solve Backend

WebSocket server for collaborative math problem solving with PostgreSQL persistence.

## Database Migrations with Atlas

This project uses [Atlas](https://atlasgo.io/) for database schema management.

### Install Atlas

```bash
curl -sSf https://atlasgo.sh | sh
```

### Local Development

```bash
# Generate migration files from schema
atlas migrate diff initial --env local

# Apply migrations
export DATABASE_URL='postgres://alpha_solve:password@localhost:5432/alpha_solve?sslmode=disable'
atlas migrate apply --env prod

# Check migration status
atlas migrate status --env prod
```

### Schema Changes

1. Edit `schema.sql` with your changes
2. Generate migration: `atlas migrate diff <name> --env local`
3. Review the generated migration in `migrations/`
4. Apply: `atlas migrate apply --env prod`
5. Commit both `schema.sql` and `migrations/` to git

### Production

Migrations run automatically during deployment. The deploy script:
1. Installs Atlas if not present
2. Runs `atlas migrate apply` with your DATABASE_URL
3. Starts the application

## Environment Variables

Create `/opt/alpha-solve/.env` on your EC2 instance:

```bash
DATABASE_URL=postgres://postgres:your_password@neelemanetdb.cwv4ig2icg26.us-east-1.rds.amazonaws.com:5432/alpha_solve?sslmode=require
```

For local development, create `.env` in the project root (not committed to git)

## Deployment

Push to `main` branch and GitHub Actions will automatically deploy to EC2.

### Required GitHub Secrets

- `EC2_HOST` - Your EC2 public IP
- `EC2_USER` - Set to `ec2-user`
- `EC2_SSH_KEY` - Your EC2 private key (.pem file contents)

## Local Development

```bash
# Run server
go run .

# With database
DATABASE_URL='postgres://...' go run .
```

## API

- `ws://host/` - WebSocket connection (requires `name`, `projectId`, `userId` query params)
- `GET /health` - Health check endpoint

