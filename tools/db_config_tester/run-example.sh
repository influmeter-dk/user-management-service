export USER_DB_CONNECTION_STR="todo"
export USER_DB_USERNAME="todo"
export USER_DB_PASSWORD="???"
export USER_DB_CONNECTION_PREFIX=""

export DB_TIMEOUT=10
export DB_IDLE_CONN_TIMEOUT=45
export DB_MAX_POOL_SIZE=8
export DB_DB_NAME_PREFIX="DB_TIMEOUT_TEST_"

export USE_NO_CURSOR_TIMEOUT=true

# flags:
# -generateUsers 12345 (number of users to be generated)
# -benchmark
go run main.go "$@"