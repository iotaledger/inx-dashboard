package app

import (
	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/core/shutdown"
	"github.com/iotaledger/hive.go/app/plugins/profiling"
	"github.com/iotaledger/inx-app/inx"
	"github.com/iotaledger/inx-dashboard/core/dashboard"
	"github.com/iotaledger/inx-dashboard/plugins/prometheus"
)

var (
	// Name of the app.
	Name = "inx-dashboard"

	// Version of the app.
	Version = "0.1.1"
)

func App() *app.App {
	return app.New(Name, Version,
		app.WithInitComponent(InitComponent),
		app.WithCoreComponents([]*app.CoreComponent{
			inx.CoreComponent,
			dashboard.CoreComponent,
			shutdown.CoreComponent,
		}...),
		app.WithPlugins([]*app.Plugin{
			profiling.Plugin,
			prometheus.Plugin,
		}...),
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
