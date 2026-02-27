# Dependency Dashboard 

## Overview
This API is using deps.dev API to store dependencies of a NPM package alongside its basic metadata and OpenSSF scores. This data is stored in SQLite database and presented in UI in form of a dependency table and OpenSSF score bar chart.

## Requirements
### Docker
- Docker
- Docker compose
### Local development
- Go 1.25+
- GCC (required by go-sqlite3)

## Local installation and running
After cloning the repo run

`docker compose up --build`
 
Application should be working on your localhost on port 8080.

`localhost:8080/`
 
Initially database is empty so you need to make PUT or POST request to add a package which dependencies you want to see. You can call an empty PUT reequest so the default package will be used ([express](https://github.com/expressjs/express)) eg.

`curl -X PUT localhost:8080/deps -H "Content-Type: application/json"`

If you want to know how to check different packages follow to the next section.

## API spec
`GET /deps`

Returns the list of the current package dependencies, it returns `text/html` or `application/json` based on request headers
```json
{
    "id": 1,
    "name": "express",
    "version": "5.2.1",
    "dependencies": [
        {
            "name":"express",
            "version":"5.2.1",
            "relation":"SELF",
            "score":8.4
        }
    ]
}
```
`GET /deps?name=body-parser&minScore=5`

GET endpoint supports strict equal filtering by name and minimum OpenSSF score filter. Response will be similar to standard endpoint but the dependencies will filtered to match query params.

`PUT /deps/{name}`

This will call deps.dev API and store dependencies of the `{name}` package in local SQLite database. Default version provided by deps.dev will be used (usually latest). You can omit the name query param and default package will be used instead. This call is idempotent, subsequent calls with the same name will just update last updated timestamp. If `PUT` will be called with different package name than that of already existing package, old package will be replaced and dependencies of new one will be returned. This endpoint supports body as well, but use one: query param or the body.
Response is the same as in `GET` endpoint.

`POST /deps`

Works the same way as `PUT` endpoint just does not support query param, pass package name through request body eg.

`curl -X POST localhost:8080/deps -H "Content-Type: application/json" -d "{\"name\": \"express\" }"`

`DELETE /deps/{name}`

Removes {name} package from the database.

## Database schema
Database consists of 2 tables: `packages` and `dependency_nodes`. `packages` stores the name, version and update timestamp of the package we want to investigate, this table should have just 1 row at all times. One to many relation connects `packages` to `dependency_nodes`, we store each dependency, alongside its metadata and OpenSSF score, to current package as a separate row. Full schema can be investigated in the repo `internal/adapter/outbound/sqlite/schema.go`. Data does not persists after container turns off 
