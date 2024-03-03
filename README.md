# Profile


## Local development

Prerequisites for local development
* DB (postgres) 
* Messaging (nats)
* Config (.env)
* Orders Service (local)


### Local DB server 

You can either create a fresh DB instance or have something more persistent where you can restore dumps from production.

To create a persistent DB:

```shell
docker run --name postgres_vh -p 5678:5432 -v pgdata_vh:/var/lib/postgresql/data -v ~/projects/vh/db_backup/:/backup -e POSTGRES_PASSWORD=password -e POSTGRES_USER=postgres -d postgres:12
```

- load a specific dump file

```shell
docker exec -it postgres_vh /bin/bash 
su postgres
createdb profiledb
pg_restore -d profiledb /backup/<your dump file>.custom
```

### NATS streaming
This one comes from `vh-srv-orders`. You'll have to run its server once. Checkout readme over there.

### Config
Ask another developer for his `.env` file

### Orders Service
Same for nats streaming above.



## OLD readme

### Prerequisites
- Have the profile database installed. You can use the initial.sql file found in the db folder.
- The `DATABASE_URL` environmental variable must be set. 
  
   `export DATABASE_URL=postgres://{db username}:{db password}@{db host}:{port}/profile`

### Integration tests
#### Steps
1. `export GO_INTEGRATION_TESTS=1`
2. Export `DATABASE_URL` variable to local postgres deployment.    (for example, on my machine this is postgres://postgres:postgres@localhost:5432/profile)
3. `go test -v -p=1 -race -timeout=60s ./...`

### Imp API Documentation

API: https://api.eurokab.info/profile/v1/profiles

Query Params:

- skip
- limit
- country
- email
- name
- ten-group-name
- language
- first-language
- other-language-1
- other-language-2
- other-language-3
- other-language-4
- phone-number
- gender
- membership ( will search as status.membership )
- membership-type ( will search as status.membership_type )
- convention ( will search as status.convention )
- ticket ( will search as status.ticket )
- galaxy ( will search as status.galaxy )
- updated ( asc || desc )
- created ( asc || desc )