#!/bin/sh
set -eu

psql --set ON_ERROR_STOP=1 \
  --username "$POSTGRES_USER" \
  --dbname postgres \
  --set identity_db="$IDENTITY_DB" \
  --set media_db="$MEDIA_DB" \
  --set identity_user="$IDENTITY_DB_USER" \
  --set identity_password="$IDENTITY_DB_PASSWORD" \
  --set media_user="$MEDIA_DB_USER" \
  --set media_password="$MEDIA_DB_PASSWORD" <<-'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'identity_user', :'identity_password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'identity_user') \gexec

SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'media_user', :'media_password')
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'media_user') \gexec

SELECT format('CREATE DATABASE %I', :'identity_db')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'identity_db') \gexec

SELECT format('CREATE DATABASE %I', :'media_db')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'media_db') \gexec

SELECT format('ALTER DATABASE %I OWNER TO %I', :'identity_db', :'identity_user') \gexec
SELECT format('ALTER DATABASE %I OWNER TO %I', :'media_db', :'media_user') \gexec
SQL
