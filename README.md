# Lakelandcup Auth Service


## Configuration

Provide a `.env` file that defines the following environment variables
```bash
HOST=
PORT=
POSTGRES_URI=
JWT_SECRET_KEY=
```

## Installation

```bash
$ make proto
```

## Running the app

```bash
# development
$ make server
```

## Connect to the Database
```bash
docker exec -it <containerhash> bin/bash
su - postgres
psql "postgresql://postgres:postgres@localhost/lakelandcup_auth_service"
```