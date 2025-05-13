# Gator

A guided project from [boot.dev](https://boot.dev/).

## Description

Gator is a command-line app to fetch RSS feeds from the internet. It's a barebones blog aggregator.

This app is written in Go and connects to PostgreSQL instances, so it should work on any popular operating system, but it was tested only on Ubuntu Linux.

## Motivation

This app was developed to practice programming in Go doing CRUD operations on a PostgreSQL instance.

## Usage

Gator only works on the command line through commands. It doesn't ship with any kind of graphical interface, not even a TUI (text-based user interface).

On a testing environment, we run the app with `go run . <command> [arguments]`.

### Available commands

These commands are available after setting up `gator`. Keep on reading this document to learn how to do that.

- `login <username>`: log in as an existing user
- `register <username>`: register a new user
- `reset`: delete all database records forever
- `users`: list all registered users
- `agg <time between requests>`: fetch the next feed stored in the database every `<time between requests>` indefinitely (time format should be human readable, like 5s500ms for 5.5 seconds)
- `addfeed <feed name> <feed URL>`: add a feed to the database and follow it
- `feeds`: list all feeds stored in the database
- `follow <url>`: follow a feed stored in the database
- `following`: list feeds followed by current user
- `unfollow <url>`: unfollow a feed followed by current user

## Requirements

- A CPU architecture and operating system compatible with the Go runtime.
- Goose to run the database migrations.
- A working installation of PostgreSQL.
- Optionally, the Go compiler to install Goose with `go install`.

### Installing Goose on Linux

Goose is written in Go, so it's trivial to install it when we have Go available on our system:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

In case we don't have and don't want to install Go itself, we can install Goose running the [following command on Linux](https://pressly.github.io/goose/installation/#linux):

```bash
curl -fsSL \
    https://raw.githubusercontent.com/pressly/goose/master/install.sh |\
    sh
```

### Installing PostgreSQL on Ubuntu Linux

We install PostgreSQL on Ubuntu running the following commands:

```bash
# update the local repository cache
sudo apt update                    # alternative: aptitude update
# update the installed packages
sudo apt upgrade                   # alternative: aptitude safe-upgrade
# install PostgreSQL
sudo apt install postgresql postgresql-contrib  # alternative: aptitude install . . .
```

We then verify that PostgreSQL was installed successfully running `psql`, the official PostgreSQL client. This client provides the user with a shell where they can run SQL commands:

```bash
# verify that PostgreSQL was installed successfully
psql --version
```

### Starting the PostgreSQL server on Ubuntu Linux

After installation, we must run the following command to start the server:

```bash
sudo service postgresql start
```

We do this once on Ubuntu, just after installing PostgreSQL and before any reboot. After a reboot, the server should start automatically. This behavior may be different on other Linux distributions.

### Log in to the PostgreSQL server through `psql` in Ubuntu Linux

On testing environments, we can log in to the local PostgreSQL server using the PostgreSQL admin user:

```bash
# run psql as the postgres user
sudo -u postgres psql
```

In case that doesn't work, we can add the following rule to `pg_hba.conf`, the configuration file of PostgreSQL:

```postgres
local all postgres peer
```

That file is usually located at `/etc/postgresql/<major version>/main/pg_hba.conf`.

We now set a password for the `postgres` user running the following query on `psql`, replacing `<new password>` with the actual password:

```sql
ALTER USER postgres PASSWORD '<new password>';
```

Notice that, with that instruction, we're changing the password for the `postgres` database user and not the Linux user.

Finally, we create the `gator` database running the following command in `psql`:

```sql
CREATE DATABASE gator;
```

We close the `psql` shell typing `exit`.

## Running the app

### Configuration

The app reads the configuration file located at `~/.gatorconfig.json` to work. This file must contain the following fields:

- `db_url`: a working connection string to a local PostgreSQL instance.
- `current_user_name`: the active user. We can omit it the first time we're running the app because `gator` will take care of it.

The connection string to the PostgreSQL database must have the following form:

```
"postgres://postgres:password@localhost:5432/gator?sslmode=disable"
```

The quotes are mandatory. We should replace `password` with the password we gave to the `postgres` user.

Also, note that `5432` is the port that PostgreSQL listens to by default. We must change that value in case our PostgreSQL instance is listening to a non-default port.

Finally, we specify `?sslmode=disable` to tell the app it shouldn't use SSL locally.

### Database migration

To migrate the `gator` database we created before, we should run the following command at the root directory of the project replacing the connection string with the actual one:

```bash
GOOSE_MIGRATION_DIR=sql/schema goose postgres "postgres://postgres:password@localhost:5432/gator" up
```

## Keep developing the code

### Requirements

- `goose` to manage database migrations.
- `sqlc` to compile SQL queries into Go code.
- A working installation of PostgreSQL.

### Install dependencies

#### Standalone programs

Install Goose and `sqlc` by running the following commands:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Here are the up-to-date installation commands in case anything goes wrong running the previous ones:

- [How to install Goose](https://github.com/pressly/goose#install).
- [How to install `sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html).

#### Libraries

Go will download and install any dependencies found in `go.mod` and `go.sum` automatically when trying to run the code with `go run . <command> [arguments]`, but we can install the required libraries explicitly by running `go mod download`.

### Migrate the database

Run the following command at the root directory of the repository to migrate the database to the state expected by `gator`:

```bash
GOOSE_MIGRATION_DIR=sql/schema goose postgres "postgres://postgres:password@localhost:5432/gator" up
```

Notice that you should replace `password` with the actual password of the `postgres` user and, eventually, also the port that PostgreSQL is listening to.

### Configure `sqlc`

The code ships with a working configuration file located at `./sqlc.yaml`.

Read `sqlc`'s [documentation](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html) to know more about how to configure that program.

### Generate Go code from SQL queries using `sqlc`

After setting up `sqlc` with `./sqlc.yaml`, we generate Go code from the queries located at `./sql/queries` running `sqlc generate` from the root directory of the repository.

## Instructor's suggestions for extending the project

- Add sorting and filtering options to the browse command.
- Add pagination to the browse command.
- Add concurrency to the agg command so that it can fetch more frequently.
- Add a search command that allows for fuzzy searching of posts.
- Add bookmarking or liking posts.
- Add a TUI that allows you to select a post in the terminal and view it in a more readable format (either in the terminal or open in a browser).
- Add an HTTP API (and authentication/authorization) that allows other users to interact with the service remotely.
- Write a service manager that keeps the agg command running in the background and restarts it if it crashes.
