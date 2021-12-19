export USER_DB_CONNECTION_STR="<db-address>"
export USER_DB_USERNAME="<db-user-name>"
export USER_DB_PASSWORD="<db-password>"
export USER_DB_CONNECTION_PREFIX="<+srv or empty>"

export DB_TIMEOUT=30
export DB_IDLE_CONN_TIMEOUT=45
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="<db name prefix if any used>"


go run main.go "$@"