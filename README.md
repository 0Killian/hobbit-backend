# Hobbit Backend

This is the backend of the Hobbit project. For more information, go to the [frontend repository](https://github.com/D-Seonay/Hobbit).

## Requirements

- Go 1.18
- PostgreSQL
- Keycloak Public Key
- Redis (optional)
- Stripe (webhook secret, secret key, 1K XP price ID)

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

## Project details

Membres:
- Killian BELLOUARD (@0Killian)
- Mathéo DELAUNAY (@D-Seonay)
- Claire NGUYEN (@podfleur)
- Sasha WILK (@jojosashaw)

Nous avons ajouté une connection optionelle à Redis pour stocker les utilisateurs authentifiés dans un cache. Nous avons également ajouté une intégration à Stripe pour acheter 1,000 points d'expérience (1 niveau).

Cette API est publiquement accessible [ici](https://hobbit-backend-mpc2.onrender.com)
