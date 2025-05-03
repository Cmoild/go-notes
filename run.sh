#!/bin/sh

echo "Waiting for PostgreSQL..."

# Используем переменные окружения из .env
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  sleep 1
done

echo "PostgreSQL is ready. Running migrations..."

# Формируем строку подключения
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Применяем миграции
psql "$DB_URL" -f internal/db/migrations/02_create_user.sql
psql "$DB_URL" -f internal/db/migrations/03_create_notes.sql
psql "$DB_URL" -f internal/db/migrations/04_create_tokens.sql

echo "Starting Go backend..."

cp /app/.env /.env

# Запускаем бэкенд
./main
