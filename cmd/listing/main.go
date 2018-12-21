package main

import (
	"fmt"
	"log"
	"os"

	heroku "github.com/heroku/heroku-go/v3"
	"github.com/shinji62/heroku-asset-listing/pkg/herokuls"
	"github.com/shinji62/heroku-asset-listing/pkg/output"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	hUsername = kingpin.Flag("heroku.username", "Heroku username").Required().Envar("HEROKU_USERNAME").String()
	hPassword = kingpin.Flag("heroku.password", "Heroku password").Required().Envar("HEROKU_PASSWORD").String()
	format    = kingpin.Flag("format", "formating output (valid values json,tab,pretty-json default to tab)").Envar("OUTPUT_FORMAT").Default("tab").Enum("json", "tab", "pretty-json")
)

const (
	ExitCodeOk    = 0
	ExitCodeError = 1 + iota
)

var (
	version     = "0.0.0"
	builddate   = ""
	commit_sha1 = ""
)

func main() {
	log.SetFlags(0)
	kingpin.Version(version)
	kingpin.Parse()

	heroku.DefaultTransport.Username = *hUsername
	heroku.DefaultTransport.Password = *hPassword

	h := heroku.NewService(heroku.DefaultClient)
	hls := herokuls.NewHerokuListing(h)
	herokuOrgs, err := hls.ListAllAppsByOrganisation()
	if err != nil {
		fmt.Println(err)
		os.Exit(ExitCodeError)
	}

	var out output.Output
	switch *format {
	case "json":
		out = output.NewJsonWriter(os.Stdout, false)
	case "pretty-json":
		out = output.NewJsonWriter(os.Stdout, true)
	case "tab":
		out = output.NewTabWriter(os.Stdout)
	default:
		fmt.Println("Only json,tab,pretty-json are accepted")
		os.Exit(ExitCodeError)
	}

	out.RenderApps(herokuOrgs)

}
