package output

import (
	"os"

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

func (t *TabWriter) RenderApps(herokuOrgs []herokuls.HerokuOrganization) {
	table := tablewriter.NewWriter(t.fileOutput)
	table.SetHeader([]string{"Name", "Released", "Updated", "Dynos", "Addons", "Stack"})

	for _, org := range herokuOrgs {
		for _, app := range org.Apps {
			status := "NOT RUNNING"
			if app.App.Name != "" {
				dynosByApp := herokuls.CountDynoTypeByApp(app.Dynos)
				appAddOns := herokuls.CountAddOnsTypeByApp(app.AddOns)
				if len(dynosByApp) > 0 {
					status = ""
				}
				table.Append([]string{app.App.Name, app.App.ReleasedAt.Format("2006-01-02"), app.App.UpdatedAt.Format("2006-01-02"), status, "", app.App.Stack.Name})
				mergedAddOnDynos := herokuls.MergeAddon(appAddOns, dynosByApp)
				for _, merge := range mergedAddOnDynos {
					table.Append([]string{"", "", "", merge[0], merge[1], ""})
				}

			}
		}

	}
	table.Render()
}
