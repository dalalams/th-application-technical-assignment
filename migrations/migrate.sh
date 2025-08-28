#!/bin/sh

echo "checking postgres..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  echo "check failed, sleeping for 2s"
  sleep 2
done

echo "postgres up"

DB_URL="postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE"

cd /migrations
goose postgres "$DB_URL" up

echo "migrations complete"
