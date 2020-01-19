export DB_CONNECTION_STR="influenzanet-dev-pwvbz.mongodb.net/test?retryWrites=true&w=majority"
export DB_USERNAME="user-management-service"
export DB_PASSWORD="89kJAO43BRUyNSbr"
export DB_PREFIX="+srv"
export DB_TIMEOUT=30
export DB_IDLE_CONN_TIMEOUT=45
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="INF_TEST_"

export USER_MANAGEMENT_LISTEN_PORT=3423
export ADDR_AUTH_SERVICE="auth-service:3423"

go test  ./...

