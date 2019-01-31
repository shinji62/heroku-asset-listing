package herokuls

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"

	heroku "github.com/heroku/heroku-go/v3"
	"go.uber.org/ratelimit"
)

// HerokuListing Listing Service for Heroku
// Overload default heroku Service
type HerokuListing struct {
	Cli *heroku.Service
	ctx context.Context
}

//HerokuOrganization Organization and Application
type HerokuOrganization struct {
	org  heroku.Organization
	Apps []HerokuApp `json:"organization_applications"`
}

//HerokuApp Heroku app with Dynos and Addon
type HerokuApp struct {
	App    heroku.OrganizationApp `json:"application"`
	Dynos  []heroku.Dyno          `json:"application_dynos"`
	AddOns []heroku.AddOn         `json:"application_addons"`
}

//DynoTypeByApp
type DynoTypeByApp struct {
	DynoSize string
	Total    int
}

//AddOnTypeByApp
type AddOnTypeByApp struct {
	Name  string
	Total int
}

func NewHerokuListing(herokuCli *heroku.Service) *HerokuListing {
	return &HerokuListing{
		Cli: herokuCli,
		ctx: context.TODO(),
	}
}

//ListAllAppsByOrganisation Aggregate all application related to an Organization
func (hls *HerokuListing) ListAllAppsByOrganisation() ([]HerokuOrganization, error) {
	organizations, err := hls.Cli.OrganizationList(hls.ctx, &heroku.ListRange{Field: "name"})
	if err != nil {
		return []HerokuOrganization{}, err
	}
	var herokuOrganisations []HerokuOrganization

	var wg = &sync.WaitGroup{}
	var mutex = &sync.Mutex{}
	errChannel := make(chan error, len(organizations))

	for _, organization := range organizations {
		wg.Add(1)
		go func(organization heroku.Organization) {
			defer wg.Done()

			apps, err := hls.getAppsbyOrg(organization)
			if err != nil {
				errChannel <- err
			}
			mutex.Lock()
			herokuOrganisations = append(herokuOrganisations, HerokuOrganization{
				org:  organization,
				Apps: apps,
			})
			mutex.Unlock()
		}(organization)
	}

	wg.Wait()
	close(errChannel)

	return herokuOrganisations, <-errChannel
}

//getAppsbyOrg Internal function which is spin up for Every Organization
//return list of Heroku App
func (hls *HerokuListing) getAppsbyOrg(organization heroku.Organization) ([]HerokuApp, error) {

	apps, err := hls.Cli.OrganizationAppListForOrganization(hls.ctx, organization.ID, &heroku.ListRange{Field: "name"})
	var herokuApps []HerokuApp
	if err != nil {
		return []HerokuApp{}, err
	}

	var wg = &sync.WaitGroup{}
	var mutex = &sync.Mutex{}
	errChannel := make(chan error, len(apps))

	// Heroku have some unclear limit on Request by sec.
	rl := ratelimit.New(40) // per second

	for _, app := range apps {
		wg.Add(1)
		go func(app heroku.OrganizationApp) {
			defer wg.Done()
			rl.Take()
			dynos, err := hls.getDynosbyApps(app)
			if err != nil {
				errChannel <- err
			}
			addOns, err := hls.getAddOnsbyApps(app)
			if err != nil {
				errChannel <- err
			}
			mutex.Lock()
			herokuApps = append(herokuApps, HerokuApp{
				App:    app,
				Dynos:  dynos,
				AddOns: addOns,
			})
			mutex.Unlock()
		}(app)
	}
	wg.Wait()
	close(errChannel)
	sort.Slice(herokuApps, func(i, j int) bool {
		return herokuApps[i].App.Name < herokuApps[j].App.Name
	})
	return herokuApps, <-errChannel
}

//getDynosbyApps List all Dynos for an application
//Spin one by app
func (hls *HerokuListing) getDynosbyApps(app heroku.OrganizationApp) ([]heroku.Dyno, error) {
	dynoArr, err := hls.Cli.DynoList(hls.ctx, app.ID, &heroku.ListRange{Field: "name"})
	var dynos []heroku.Dyno
	if err != nil || len(dynoArr) == 0 {
		return dynos, err
	}

	for _, dyno := range dynoArr {
		dynos = append(dynos, dyno)
	}
	return dynos, nil

}

func (hls *HerokuListing) getAddOnsbyApps(app heroku.OrganizationApp) ([]heroku.AddOn, error) {
	addOnArr, err := hls.Cli.AddOnListByApp(hls.ctx, app.ID, &heroku.ListRange{Field: "name"})
	var addOns []heroku.AddOn
	if err != nil {
		return addOns, err
	}
	for _, addOn := range addOnArr {
		addOns = append(addOns, addOn)
	}

	return addOns, nil

}

func (hls *HerokuListing) GetRateLimitingRemaining() (int, error) {
	rate, err := hls.Cli.RateLimitInfo(hls.ctx)
	if err != nil {
		return 0, err
	}
	return rate.Remaining, nil
}

func (hls *HerokuListing) GetDynoSizeInformation() (map[string]int, error) {
	dynosSize, err := hls.Cli.DynoSizeList(hls.ctx, &heroku.ListRange{Field: "id"})
	dynoInfo := make(map[string]int, len(dynosSize))
	if err != nil {
		return dynoInfo, err
	}
	for _, dynoSize := range dynosSize {
		dynoInfo[dynoSize.Name] = dynoSize.DynoUnits
	}
	return dynoInfo, nil
}

func CountDynoTypeByApp(dynos []heroku.Dyno) []DynoTypeByApp {
	var dynoTypeByApp []DynoTypeByApp
	dynosCumulated := CountDynosCumulated(dynos)
	for dynoSize, dynoCount := range dynosCumulated {
		dynoTypeByApp = append(dynoTypeByApp, DynoTypeByApp{
			DynoSize: dynoSize,
			Total:    dynoCount,
		})

	}
	return dynoTypeByApp
}

func CountTotalDynoUnitByApp(dynosByApp []DynoTypeByApp, dynoSize map[string]int) int {
	var totalUnitByApp int
	for _, dyno := range dynosByApp {
		totalUnitByApp += dyno.Total * dynoSize[dyno.DynoSize]
	}
	return totalUnitByApp
}

func CountDynosCumulated(dynos []heroku.Dyno) map[string]int {
	dynoType := make(map[string]int)
	if len(dynos) > 0 {
		for _, dyno := range dynos {
			if _, ok := dynoType[dyno.Size]; ok {
				dynoType[dyno.Size]++
			} else {
				dynoType[dyno.Size] = 1
			}
		}
	}
	return dynoType
}

func CountAddOnsTypeByApp(addOns []heroku.AddOn) []AddOnTypeByApp {
	var addOnsTypeByApp []AddOnTypeByApp
	addOnsCumulated := CountAddOnsCumulated(addOns)
	for addOnName, addOnCount := range addOnsCumulated {
		addOnsTypeByApp = append(addOnsTypeByApp, AddOnTypeByApp{
			Name:  addOnName,
			Total: addOnCount,
		})

	}
	return addOnsTypeByApp
}

func CountAddOnsCumulated(addons []heroku.AddOn) map[string]int {
	addOnsType := make(map[string]int)
	if len(addons) > 0 {
		for _, addOn := range addons {
			if _, ok := addOnsType[addOn.AddonService.Name]; ok {
				addOnsType[addOn.AddonService.Name]++
			} else {
				addOnsType[addOn.AddonService.Name] = 1
			}
		}
	}
	return addOnsType
}

func MergeAddon(addOns []AddOnTypeByApp, dynos []DynoTypeByApp) [][]string {
	var mergedString [][]string
	if len(addOns)+len(dynos) == 0 {
		return mergedString
	}
	max := 0
	if len(addOns) > len(dynos) {
		max = len(addOns)
	} else {
		max = len(dynos)
	}
	for i := 0; i <= max; i++ {
		var dyno string
		var addOn string
		if i < len(dynos) {
			dyno = dynos[i].DynoSize + " " + strconv.Itoa(dynos[i].Total)
		}
		if i < len(addOns) {
			addOn = addOns[i].Name + " " + strconv.Itoa(addOns[i].Total)
		}
		mergedString = append(mergedString, []string{dyno, addOn})
	}

	return mergedString
}

// GetIPList get all ips from all spaces of the user's enterprise teams
func (hls *HerokuListing) GetIPList(name, description string) *IPList {
	ts, err := hls.Cli.TeamList(hls.ctx, &heroku.ListRange{Field: "id"})
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
	spaces, err := hls.GetSpacesFromTeams(&teams)
	if err != nil {
		return nil
	}
	ipList, err := hls.buildIPListFromSpaces(name, description, &spaces)
	if err != nil {
		return nil
	}
	return ipList
}

// GetSpacesFromTeams get spaces that the provided teams own
func (hls *HerokuListing) GetSpacesFromTeams(ts *[]heroku.Team) ([]heroku.Space, error) {
	teams := map[string]bool{}
	for _, team := range *ts {
		teams[team.ID] = true
	}

	spaces, err := hls.Cli.SpaceList(hls.ctx, &heroku.ListRange{Field: "id"})
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
func (hls *HerokuListing) buildIPListFromSpaces(name, description string, spaces *[]heroku.Space) (*IPList, error) {
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

			spaceNat, err := hls.Cli.SpaceNATInfo(hls.ctx, space.ID)
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
