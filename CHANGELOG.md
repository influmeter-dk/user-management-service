# Changelog

## [v0.19.3] - 2021-06-03

### Added:

- New tool to create an admin user. This is located in [here](tools/create-admin-user)

### Changed:

- Include user role "service account", when looking up non-participant users.
- Adding context for timer event's run method, to prepare logic for graceful shutdown.
- gRPC endpoint for creating a new user (`CreateUser`), accept a list of profile names that is then used to create profiles. The first profile name will be assigned to the main profile. If the list is empty, the blurred email address will be used for the main profile as before.
