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
	cli       = kingpin.New("heroku-listing", "assets listing tool from devops")
	hUsername = cli.Flag("heroku.username", "Heroku username").Required().Envar("HEROKU_USERNAME").String()
	hPassword = cli.Flag("heroku.password", "Heroku password").Required().Envar("HEROKU_PASSWORD").String()
	hToken    = cli.Flag(
		"heroku.token",
		"(Optional) Heroku Authorizations Token. If token is present, basic auth will be ignored.",
	).Short('t').Envar("HEROKU_AUTH_TOKEN").String()

	cloud         = cli.Command("cloud", "list cloud assets")
	format        = cloud.Flag("format", "formating output (valid values json,tab,pretty-json default to tab)").Envar("OUTPUT_FORMAT").Default("tab").Enum("json", "tab", "pretty-json")
	dynoUnitPrice = cloud.Flag("heroku.dyno-unit-price", "Price in $ of 1 dyno unit (default 0)").Envar("HEROKU_DYNO_PRICE").Default("0").Int()

	ips        = cli.Command("ips", "list cloud assets")
	outputFile = ips.Flag("output", "Output filename").Short('o').Default("ips-listing.yml").String()
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
	cli.Version(version)

	cmd := kingpin.MustParse(cli.Parse(os.Args[1:]))
	heroku.DefaultTransport.Username = *hUsername
	heroku.DefaultTransport.Password = *hPassword
	if hToken != nil {
		heroku.DefaultTransport.BearerToken = *hToken
	}

	h := heroku.NewService(heroku.DefaultClient)
	hls := herokuls.NewHerokuListing(h)

	switch cmd {
	case cloud.FullCommand():
		herokuOrgs, err := hls.ListAllAppsByOrganisation()
		if err != nil {
			fmt.Println(err)
			os.Exit(ExitCodeError)
		}

		dynoSize, err := hls.GetDynoSizeInformation()
		if err != nil {
			fmt.Println(err)
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

		out.RenderApps(herokuOrgs, dynoSize, *dynoUnitPrice)
	case ips.FullCommand():
		ipList := hls.GetIPList("heroku-ips-listing", "ips from heroku spaces")

		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error opening file: %v", err))
			return
		}
		defer f.Close()
		ipList.Yamlize(f)
		fmt.Println(fmt.Sprintf("Success! Created file: %s", *outputFile))
	}

}
