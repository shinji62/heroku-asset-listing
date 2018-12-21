package output

import (
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/shinji62/heroku-asset-listing/pkg/herokuls"
)

type JsonWriter struct {
	pretty bool
	file   *os.File
}

func NewJsonWriter(fileOutput *os.File, prettyJson bool) *JsonWriter {
	return &JsonWriter{
		pretty: prettyJson,
		file:   fileOutput,
	}
}

func (j *JsonWriter) RenderApps(herokuOrgs []herokuls.HerokuOrganization) {

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var b []byte
	var err error
	if j.pretty {
		b, err = json.MarshalIndent(herokuOrgs, "", "  ")
	} else {
		b, err = json.Marshal(herokuOrgs)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(j.file, "%s", b)

}
