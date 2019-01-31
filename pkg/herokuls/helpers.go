package herokuls

import (
	"fmt"
	"io"

	yaml "gopkg.in/yaml.v2"
)

// IPList root, contains descriptive name and array of IP items
type IPList struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	IPListItems []IPListItem `yaml:"items"`
}

// IPListItem an IP List Item
type IPListItem struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	IPList      []string `yaml:"ips"`
}

const (
	// TeamTypeEnterprise Team.Type is enterprise
	TeamTypeEnterprise = "enterprise"
)

// Yamlize write to file as yaml
func (ipList *IPList) Yamlize(w io.Writer) error {
	en := yaml.NewEncoder(w)
	defer en.Close()
	if err := en.Encode(ipList); err != nil {
		fmt.Print(err)
		return err
	}
	return nil
}
