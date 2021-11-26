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