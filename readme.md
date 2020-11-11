# User management service for the Influenzanet system

This is a Go implementation of the [User Management Service](https://github.com/influenzanet/influenzanet/wiki/Services#user-management-service)

It provides operations to manage User accounts and profiles for an InfluenzaNet platform.


## Test
Before running the test first you have to generate the client mock services:
```
make mock
```
This assumes that the other services (messaging-service) is in the same parent folder as this package.

With a running go setup, you can use the command
```
make test
```
to execute the test script. Makefile expects the test script to be at test/test.sh. The test script could contain DB secrets therefore are not added to this git repository. An example [test script](test/example_test_srcipt.sh) can be found in the `test` folder.

Currently the tests also require a working database connection to a mongoDB instance.

## Build
### Docker
Dockerfile(s) are located in `build/docker`. The default Dockerfile is using a multistage build and create a minimal image base on `scratch`.
To trigger the build process using the default docker file call:
```
make docker
```
This will use the most recent git tag to tag the docker image.

#### Contribute your deployment setup:
Feel free to create your own Dockerfile (e.g. compiling and deploying to specific target images), eventually others may need the same.
You can create a pull request with adding the Dockerfile into `build/docker` with a good name that it can be identified well, and add a short description to `build/docker/readme.md` about the purpose and speciality of it.

An example to run your created docker image - with the set environment variables - can be found [here](build/docker/example).

## Config

The available environment variables to configure the services are available in the docker example file.

### JWT_TOKEN_KEY
The private key JWT_TOKEN_KEY can be generated using the `key-generator` tool provided. It obviously needs to be stored in a secured way once generated.

## Misc
Maximum ten devices can get a refresh token at the same time - see pkg/models/user.go

Create a database index on user collection for:
- account.accountID
- account.accountConfirmedAt + timestamps.createdAt
