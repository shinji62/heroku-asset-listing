package output

import "github.com/shinji62/heroku-asset-listing/pkg/herokuls"

type Output interface {
	RenderApps(herokuOrgs []herokuls.HerokuOrganization)
}
