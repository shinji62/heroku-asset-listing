# Description

Nifty tool which collect data from Heroku using Heroku API.
Support three types of output format
* Tab Mainly for presenting
* json
* pretty Json

This tool use Go modules and is compiled with Go 1.11.X

## Compile
```
go build -o heroku-listing cmd/listing/main.go
```

## Usage
Just run `forwarder --help` to get the latest help

```bash
usage: heroku-listing --heroku.username=HEROKU.USERNAME --heroku.password=HEROKU.PASSWORD [<flags>] <command> [<args> ...]

assets listing tool from devops

Flags:
      --help     Show context-sensitive help (also try --help-long and --help-man).
      --heroku.username=HEROKU.USERNAME
                 Heroku username
      --heroku.password=HEROKU.PASSWORD
                 Heroku password
  -t, --heroku.token=HEROKU.TOKEN
                 (Optional) Heroku Authorizations Token. If token is present, basic auth will be ignored.
      --version  Show application version.

Commands:
  help [<command>...]
    Show help.

  cloud [<flags>]
    list cloud assets
```
## Environment Variable
This application support Environment

* Heroku username `HEROKU_USERNAME`
* Heroku password `HEROKU_PASSWORD`
* Heroku token `HEROKU_AUTH_TOKEN`

* Format `OUTPUT_FORMAT`


## Build

```
go build -o heroku-listing cmd/listing/main.go '-mod=vendor'
```

## Test ..... No yet implemented
```
go test ./... -mod=vendor -v
```


## Todo
  [] Test
