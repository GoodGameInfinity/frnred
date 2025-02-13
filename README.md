# Frnred

Frnred (**Fr**ie**n**d's **red**irect) is a barebones URL shortener written in Go with Fiber.
It's designed specifically for the frnsrv.ru website.

## Installation
### Go
```bash
go install go.frnsrv.ru/frnred@latest
```
### Container image
Currently there is no official pre-built image.

## Building
### Go
```bash
go build .
```
### Container image
Tested with Podman but should work with Docker too
```bash
podman image build -t gginfinity/frnred .
```

## Usage & Configuration
Frnred is configured through CLI flags passed on execution. The default config (when you don't specify anything) looks like this:
```bash
-addr "0.0.0.0:8080" -db "file:./frnred.db" -root "https://friendsserver.ru"
```

To use this shortener you need a database. Currently frnred supports libsql, postgres, and mysql-compatable DBs.
Before using a DB please add the schema from query/schema.sql.
See the -db flag help info for DB connection setup.

frnred -help:

[//]: # (BEGIN HELPINFO)
```bash
Usage: frnred [options]
Options:
  -addr string
    	Application address string (0.0.0.0:8080) (default "0.0.0.0:8080")
  -db string
    	Database connection string (file: or libsql:// for libsql | postgres:// | | sql:// etc.) (default "file:./frnred.db")
  -help
    	Shows this message
  -root string
    	Root URL redirect (https://friendsserver.ru) (default "https://friendsserver.ru")

This program is a URL shortener designed for frnsrv.ru.
Licensed under the MIT-0, copyright 2025 GoodGameInfinity.
```
[//]: # (END HELPINFO)

[//]: # (For contributors: make sure that this thing is always up to date. Thank you in advance)
[//]: # (There even is a script that does it for you!!! Just run ./.scripts/update-help.sh or use pre-commit)

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT No Attribution](https://choosealicense.com/licenses/mit-0/)
