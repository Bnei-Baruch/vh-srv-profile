## Profile
### Prerequisites
- Have the profile database installed. You can use the initial.sql file found in the db folder.
- The `DATABASE_URL` environmental variable must be set. 
  
   `export DATABASE_URL=postgres://{db username}:{db password}@{db host}:{port}/profile`

### Integration tests
#### Steps
1. `export GO_INTEGRATION_TESTS=1`
2. Export `DATABASE_URL` variable to local postgres deployment.    (for example, on my machine this is postgres://postgres:postgres@localhost:5432/profile)
3. `go test -v -p=1 -race -timeout=60s ./...`
