# User management service for the Influenzanet system

This is a Go implementation of the [User Management Service](https://github.com/influenzanet/influenzanet/wiki/Services#user-management-service)

It provides operations to manage User accounts and profiles for an InfluenzaNet platform

It follows [common Go organization](https://github.com/influenzanet/influenzanet/wiki/Go-based-service-organisation)

## TODO API docs (link)

## Dev operations

Make targets :

 - build: build service locally
 - install-dev : install dev dependencies (to make & run test)
 - docker: build docker image
 - test: run test suites

## Test

install dev dependencies using `make install-dev`

Then generate mock client for the authentication-service:

```sh
mockgen -source=./api/auth-service-api.pb.go AuthServiceApiClient > mocks/auth-service-api.go
```

For more information about testing grpc clients with go check: <https://github.com/grpc/grpc-go/blob/master/Documentation/gomock-example.md>

## To sort

Maximum ten devices can get a refresh token at the same time - see models_user.go