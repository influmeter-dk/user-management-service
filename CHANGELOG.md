# Changelog

## [v0.20.2] - 2021-07-27

### Security Update:

- Migrating to `github.com/golang-jwt/jwt`
- Updating other dependencies

## [v0.20.1] - 2021-07-01

### Changed:

- LoginWithExternalIDP: user newly created user object to handle first time login.


## [v0.20.0] - 2021-07-01

### Added:

- New endpoint: LoginWithExternalIDP. This method handles logic for login process when a user is using an external identity provider (IDP) to login. If user did not exist in the system before, an account with type "external" will be created. If an account of type "email" already exists, the method will fail.

### Changed:

- LoginWithEmail endpoint will check account type, if external account is accessed through this endpoint, login will fail - use the external IDP instead.
- minor code improvements to use globally defined constants instead of locally hard-coded strings

## [v0.19.4] - 2021-06-16

### Changed:

- Changing endpoint for auto verificiation code generation through temp token received by email. There were occasional reports of people not able to login with email link. After catching one of such instances, it is likely that somehow a double request to that endpoint caused the replacement of the verification code. With this update, if the user identified by temp token, has a recently generated valid verification code in the DB, we won't replace it, but send this one back (agian).


## [v0.19.3] - 2021-06-03

### Added:

- New tool to create an admin user. This is located in [here](tools/create-admin-user)

### Changed:

- Include user role "service account", when looking up non-participant users.
- Adding context for timer event's run method, to prepare logic for graceful shutdown.
- gRPC endpoint for creating a new user (`CreateUser`), accept a list of profile names that is then used to create profiles. The first profile name will be assigned to the main profile. If the list is empty, the blurred email address will be used for the main profile as before.
