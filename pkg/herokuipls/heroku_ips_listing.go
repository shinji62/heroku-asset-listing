package herokuipls

import (
	"context"
	"fmt"
	"io"
	"sync"

	heroku "github.com/heroku/heroku-go/v3"
	"go.uber.org/ratelimit"
	yaml "gopkg.in/yaml.v2"
)

type (
	// YAML formatting

	// IPList root, contains descriptive name and array of IP items
	IPList struct {
		Name        string       `yaml:"name"`
		Description string       `yaml:"description"`
		IPListItems []IPListItem `yaml:"items"`
	}

	// IPListItem an IP List Item
	IPListItem struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		IPList      []string `yaml:"ips"`
	}

	// END YAML formatting

	// HerokuService heroku service bootstrap
	HerokuService struct {
		Svc *heroku.Service
		ctx context.Context
	}
)

const (
	// TeamTypeEnterprise Team.Type is enterprise
	TeamTypeEnterprise = "enterprise"
)

// NewHerokuService creates a new heroku service bootstrap
func NewHerokuService(hs *heroku.Service) *HerokuService {
	return &HerokuService{
		Svc: hs,
		ctx: context.Background(),
	}
}

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

// GetIPList get all ips from all spaces of the user's enterprise teams
func (hs *HerokuService) GetIPList(name, description string) *IPList {
	ts, err := hs.Svc.TeamList(hs.ctx, &heroku.ListRange{Field: "id"})
	if err != nil {
		fmt.Println(fmt.Sprintf("Error on TeamList: %v", err))
	}
	// get only enterprise teams
	var teams []heroku.Team
	for _, team := range ts {
		if team.Type == TeamTypeEnterprise {
			teams = append(teams, team)
		}
	}
	spaces, err := hs.GetSpacesFromTeams(&teams)
	if err != nil {
		return nil
	}
	ipList, err := hs.buildIPListFromSpaces(name, description, &spaces)
	if err != nil {
		return nil
	}
	return ipList
}

// GetSpacesFromTeams get spaces that the provided teams own
func (hs *HerokuService) GetSpacesFromTeams(ts *[]heroku.Team) ([]heroku.Space, error) {
	teams := map[string]bool{}
	for _, team := range *ts {
		teams[team.ID] = true
	}

	spaces, err := hs.Svc.SpaceList(hs.ctx, &heroku.ListRange{Field: "id"})
	if err != nil {
		fmt.Println(fmt.Sprintf("Error on SpaceList: %v", err))
		return nil, err
	}
	var res []heroku.Space
	for _, space := range spaces {
		if _, exists := teams[space.Team.ID]; exists {
			res = append(res, space)
		}
	}
	return res, nil
}

// build an IPList instances using heroku.Space info
func (hs *HerokuService) buildIPListFromSpaces(name, description string, spaces *[]heroku.Space) (*IPList, error) {
	if spaces == nil { // save the dereference
		return nil, nil
	}

	ipList := &IPList{
		Name:        name,
		Description: description,
	}
	var wg sync.WaitGroup
	var mutex sync.Mutex
	errChan := make(chan error)

	rl := ratelimit.New(40) // per second

	for _, s := range *spaces {
		wg.Add(1)
		go func(space heroku.Space) {
			defer wg.Done()
			rl.Take()

			spaceNat, err := hs.Svc.SpaceNATInfo(hs.ctx, space.ID)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on SpaceNATInfo during buildIPListFromSpaces: %v", err))
				errChan <- err
				return
			}

			mutex.Lock()
			ipList.IPListItems = append(ipList.IPListItems, IPListItem{
				Name:        fmt.Sprintf("%s/%s", space.Team.Name, space.Name),
				Description: fmt.Sprintf("IP list from `%s > %s`", space.Team.Name, space.Name),
				IPList:      spaceNat.Sources,
			})
			mutex.Unlock()
		}(s)
	}

	wg.Wait()
	close(errChan)
	return ipList, <-errChan
}
