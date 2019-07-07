# User management service for the Influenzanet system
## TODO: General description what this service is doing

## TODO API docs (link)


## Test
The tests use the gomock library. To install this use:
```
go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
```

Then generate mock client for the authentication-service:
```
mockgen -source=./api/auth-service.pb.go AuthServiceApiClient > mocks/auth-service-api.go
```
For more information about testing grpc clients with go check: https://github.com/grpc/grpc-go/blob/master/Documentation/gomock-example.md
