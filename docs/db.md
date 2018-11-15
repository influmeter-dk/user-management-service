# Database connection
## Setup
This service connects to a MongoDB instance using username and password authentication.
The database address can be configured by changing the `db_address` value in the `configs.yaml` file. Username and password for authenticating with the database are stored also in a Yaml file with the content:

```yaml
username: <db-user>
password: <db-password>
```

After a successful connection to the running mongo instance is established, the service will assume the presence of the `users` database with a `users` collection.
