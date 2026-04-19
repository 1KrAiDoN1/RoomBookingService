#!/bin/bash
set -e

# Дефолтное значение DB_URL
DB_URL="${DB_URL:-postgres://postgres:admin@postgres:5432/room_booking?sslmode=disable}"

echo "Running migrations..."
echo "Database: ${DB_URL}"


# Запускаем миграции
migrate -path ./migrations -database "${DB_URL}" -verbose up