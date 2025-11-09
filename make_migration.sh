#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: ./make_migration.sh <migration_name>"
    echo "Example: ./make_migration.sh add_users_table"
    exit 1
fi

atlas migrate diff "$1" --env local

echo "✓ Migration created. Review it in migrations/ then apply with:"
echo "  atlas migrate apply --env local"

