# Hobbit Backend

This is the backend of the Hobbit project.

## Requirements

- Go 1.18
- PostgreSQL
- Keycloak Public Key

## Usage

Put your public key in the path denoted by the `KEYCLOAK_PUBLIC_KEY_PATH` environment variable (see [.env.sample](.env.sample)).

Run the migrations: connect on your database and run schemas script in this order:
- user.sql
- task.sql

Run the server:

```bash
go mod tidy
go run main.go
```
