package output

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/shinji62/heroku-asset-listing/pkg/herokuls"
)

type TabWriter struct {
	fileOutput *os.File
}

func NewTabWriter(output *os.File) *TabWriter {
	return &TabWriter{
		fileOutput: output,
	}
}

func formatPrice(totalDynosUnit int, dynoUnitPrice int) string {
	if totalDynosUnit == 0 {
		return ""
	}
	return strconv.Itoa(totalDynosUnit) + " (" + strconv.Itoa(totalDynosUnit*dynoUnitPrice) + "$)"
}

func (t *TabWriter) RenderApps(herokuOrgs []herokuls.HerokuOrganization, dynoSize map[string]int, dynoUnitPrice int) {
	table := tablewriter.NewWriter(t.fileOutput)
	table.SetHeader([]string{"Name", "Released", "Updated", "Dynos", "d.units", "Addons", "Stack"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCaption(true, "Price by dyno is "+strconv.Itoa(dynoUnitPrice)+" a month. Total price is for a full time running dyno.")
	table.SetCenterSeparator("|")
	for _, org := range herokuOrgs {
		for _, app := range org.Apps {
			status := "NOT RUNNING"
			if app.App.Name != "" {
				dynosByApp := herokuls.CountDynoTypeByApp(app.Dynos)
				appAddOns := herokuls.CountAddOnsTypeByApp(app.AddOns)
				price := formatPrice(herokuls.CountTotalDynoUnitByApp(dynosByApp, dynoSize), dynoUnitPrice)
				if len(dynosByApp) > 0 {
					status = ""
				}
				table.Append([]string{app.App.Name, app.App.ReleasedAt.Format("2006-01-02"), app.App.UpdatedAt.Format("2006-01-02"), status, price, "", app.App.Stack.Name})
				mergedAddOnDynos := herokuls.MergeAddon(appAddOns, dynosByApp)
				for _, merge := range mergedAddOnDynos {
					table.Append([]string{"", "", "", merge[0], "", merge[1], ""})
				}

			}
		}

	}
	table.Render()
}
