# CLOX
Welcome to CLOX (CLoud bOX) - the cutting-edge cloud-based file storage service 
where security meets convenience. CLOX ensures that your files are encrypted 
end-to-end, meaning they are encrypted on your client device and remain encrypted 
until they are decrypted on the client side. Our servers never access your files 
in an unencrypted state and do not hold the keys to decrypt your files.

## Key Features
- **End-to-End Encryption**: Every file you store on CLOX is encrypted on your device 
before it even reaches our server. 
- **Secure Storage**: Rest easy knowing your data is stored securely in the cloud,
with robust protections against unauthorized access or breaches. Even if our server
is breached, your files are encrypted at rest.
- **OAuth2 Integration for Secure Sign-In**: CLOX leverages OAuth2 providers to offer
a secure and convenient sign-in experience. This means you can log in using your 
existing accounts with trusted providers, enhancing security while simplifying access.
- **Client Side Keys**: Private keys are stored on the client side. This means only 
you can decrypt and access your files - not even CLOX has the keys.
- **CLI Client**: CLOX features a command-line interface (CLI) client application, 
making it ideal for those who prefer or require a scriptable, terminal-based 
interaction.
- **API Token Authentication**: For interactions via the CLI application, users 
generate an API token through the server-side application. This token is used for 
authenticating API requests, ensuring secure and streamlined access to your files 
without compromising security.

## Local Development
This section documents the configuration to get up and running locally. 

**Important: All commands are being executed on bash/zsh shell. Some of the commands will not
work on Windows machines.**

### Requirements
- [Go](https://go.dev/dl/)
- [Docker](https://www.docker.com/get-started/)
- [Migrate](https://github.com/golang-migrate/migrate) (golang-migrate)
- [Google Account](https://console.cloud.google.com) (Google Cloud for OAuth2)

### Environment Variables
CLOX makes use of a `env` file. The `.env.template` file provides a template. To configure the `env`
file, make a copy of `.env.template` and save it as `.env`. For full functionality all variables 
must be given a valid value. 

To be consistent with the documentation, these variables should be set to the exact values:

| Environment Variable | Value         |
|----------------------|---------------|
| APP_ENV              | dev           |
| HOST                 | localhost     |
| PORT                 | 8080          |
| API_PORT             | 8081          |
| POSTGRES_HOST        | localhost     |
| POSTGRES_PORT        | 5432          |
| POSTGRES_USERNAME    | clox          |
| POSTGRES_PASSWORD    | cloxpassword  |
| POSTGRES_DBNAME      | clox_db       |
| REDIS_HOST           | localhost     |
| REDIS_PORT           | 6379          |
| REDIS_USERNAME       | clox          |
| REDIS_PASSWORD       | cloxpassword  |
| FILE_STORE_PATH      | ./dev-storage |

The `FILE_STORE_PATH` is not used in the documentation, however, please use `./dev-storage` during local development to eliminate congestion in the `.gitignore` file.

The following environment variables will be generated when creating your OAuth2 credentials, set these accordingly:

| Environment Variable       |
|----------------------------|
| GOOGLE_OAUTH_CLIENT_ID     |
| GOOGLE_OAUTH_CLIENT_SECRET |

The remaining environment variables need to be set to any value:

| Environment Variable |
|----------------------|
| JWT_SECRET_KEY       |

### Google OAuth2
Create a new project in the [Google Cloud Console](https://console.cloud.google.com/) and name it `clox`.

Once created, navigate to `API & Services`, then to `OAuth consent screen`. Select `External` and create the
consent screen. 

For `App name` enter `clox`.

Select your email for `User support email`.

Enter your email for the `Developer contact information`, then click `Save and Continue`.

In the `Scopes` section, add the following scopes then click `Save and Continue`:
- `/auth/userinfo.profile`
- `/auth/userinfo.email`
- `openid`

Since the application will have a publishing status of `Testing` by default, you must specify
what users will be able to access the application while in the `Testing` status. Add all the Google
accounts that will be accessing the application.

Click `Save and Continue` then `Back to Dashboard`.

Select the `Credentials` tab and then click `Create Credentials`. 

Choose `OAuth client ID`.

For the `Application Type` select `Desktop app` and set the name to `clox server`.

Finally, you are presented with the `Client ID` and `Client Secret`. You can download a JSON file 
that will include these values.

In your `.env` file, set `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` to these values.

**The Google OAuth client ID and client secret are not limited to local development. These values 
may be used in production. No one should have access to these values besides you.**

### Postgres
Create the Postgres container:

`docker create --name clox_postgres -e POSTGRES_DB=clox_db -e POSTGRES_USER=clox
-e POSTGRES_PASSWORD=cloxpassword -p 5432:5432 postgres:15.1-alpine`

Start the container:

`docker start clox_postgres`

The Postgres container will be started and bound to the local port `5432`.

If you want to enter the Postgres console to manage the `clox_db` database you 
can use the following:

`docker exec -it clox_postgres psql -U clox -d clox_db`

### Redis
Redis authentication is configured with `redis.conf` and `users.acl` in the `docker` directory. These
files set the username to `clox`, the password to `cloxpassword`, and grants this user all permissions.
Also, the default user is disabled.

**The `docker` directory is used for local development only. `redis.conf` and `users.acl` configure
the Redis container that is ran in local development.**

Create the Redis container: 

`docker create --name clox_redis -v $PWD/docker/redis.conf:/usr/local/etc/redis/redis.conf 
-v $PWD/docker/users.acl:/etc/redis/users.acl -p 6379:6379 redis redis-server 
/usr/local/etc/redis/redis.conf`

Start the container:

`docker start clox_redis`

The Redis container will be started and bound to the local port `6379`.

If you want to enter the Redis console to manage Redis, use the following command:

`docker exec -it clox_redis redis-cli -u redis://clox:cloxpassword@localhost:6379`

### Migrate
Use the following command to create new database migrations:

`migrate create -ext sql -dir migrations -seq <migration-name>`

To migrate up in the Postgres container, run the following command:

`migrate -path ./migrations -database "postgres://clox:cloxpassword@localhost:5432/clox_db?sslmode=disable" up`

To migrate down in the Postgres container, run the following command:

`migrate -path ./migrations -database "postgres://clox:cloxpassword@localhost:5432/clox_db?sslmode=disable" down`