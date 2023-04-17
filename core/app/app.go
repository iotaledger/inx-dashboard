package app

import (
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/components/profiling"
	"github.com/iotaledger/hive.go/app/components/shutdown"
	"github.com/iotaledger/inx-app/core/inx"
	"github.com/iotaledger/inx-dashboard/core/dashboard"
	"github.com/iotaledger/inx-dashboard/plugins/prometheus"
)

var (
	// Name of the app.
	Name = "inx-dashboard"

	// Version of the app.
	Version = "1.0.0-rc.5"
)

func App() *app.App {
	return app.New(Name, Version,
		app.WithInitComponent(InitComponent),
		app.WithComponents(
			inx.Component,
			dashboard.Component,
			shutdown.Component,
			profiling.Component,
			prometheus.Component,
		),
	)
}

var (
	InitComponent *app.InitComponent
)

func init() {
	InitComponent = &app.InitComponent{
		Component: &app.Component{
			Name: "App",
		},
		NonHiddenFlags: []string{
			"config",
			"help",
			"version",
		},
	}
}
