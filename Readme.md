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
uusage: heroku-listing --heroku.username=HEROKU.USERNAME --heroku.password=HEROKU.PASSWORD [<flags>]

Flags:
  --help        Show context-sensitive help (also try --help-long and --help-man).
  --heroku.username=HEROKU.USERNAME
                Heroku username
  --heroku.password=HEROKU.PASSWORD
                Heroku password
  --format=tab  formating output (valid values json,tab,pretty-json default to tab)
  --version     Show application version.

```
## Environment Variable
This application support Environment

* Heroku username `HEROKU_USERNAME`
* Heroku password `HEROKU_PASSWORD`
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
