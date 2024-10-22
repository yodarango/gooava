#!/bin/bash

# Database connection details (adjust these according to your setup)
DB_USER="your_username"
DB_PASSWORD="your_password"
DB_NAME="your_database_name"

# Run each migration file in order
echo "Running migrations..."

mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/migrations/001_create_users_table.sql
if [ $? -ne 0 ]; then
    echo "Migration 001 failed!"
    exit 1
fi

mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/migrations/002_create_recipes_table.sql
if [ $? -ne 0 ]; then
    echo "Migration 002 failed!"
    exit 1
fi

mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/migrations/003_create_prompts_table.sql
if [ $? -ne 0 ]; then
    echo "Migration 003 failed!"
    exit 1
fi

mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/migrations/004_create_ingredients_table.sql
if [ $? -ne 0 ]; then
    echo "Migration 004 failed!"
    exit 1
fi

mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/migrations/005_create_recipe_ingredients_table.sql
if [ $? -ne 0 ]; then
    echo "Migration 005 failed!"
    exit 1
fi

echo "All migrations applied successfully!"
