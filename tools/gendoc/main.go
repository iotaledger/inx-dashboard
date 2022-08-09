package main

import (
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/apputils/config"
	"github.com/iotaledger/hive.go/core/app"

	dashboardApp "github.com/iotaledger/inx-dashboard/core/app"
)

func createMarkdownFile(app *app.App, markdownHeaderPath string, markdownFilePath string, ignoreFlags map[string]struct{}, replaceTopicNames map[string]string) {

	var markdownHeader []byte

	if markdownHeaderPath != "" {
		var err error
		markdownHeader, err = os.ReadFile(markdownHeaderPath)
		if err != nil {
			panic(err)
		}
	}

	println(fmt.Sprintf("Create markdown file for %s...", app.Info().Name))
	md := config.GetConfigurationMarkdown(app.Config(), app.FlagSet(), ignoreFlags, replaceTopicNames)
	os.WriteFile(markdownFilePath, append(markdownHeader, []byte(md)...), 0644)
	println(fmt.Sprintf("Markdown file for %s stored: %s", app.Info().Name, markdownFilePath))
}

func createDefaultConfigFile(app *app.App, configFilePath string, ignoreFlags map[string]struct{}) {
	println(fmt.Sprintf("Create default configuration file for %s...", app.Info().Name))
	conf := config.GetDefaultAppConfigJSON(app.Config(), app.FlagSet(), ignoreFlags)
	os.WriteFile(configFilePath, []byte(conf), 0644)
	println(fmt.Sprintf("Default configuration file for %s stored: %s", app.Info().Name, configFilePath))
}

func main() {

	// MUST BE LOWER CASE
	ignoreFlags := make(map[string]struct{})

	replaceTopicNames := make(map[string]string)
	replaceTopicNames["app"] = "Application"
	replaceTopicNames["inx"] = "INX"

	application := dashboardApp.App()

	createMarkdownFile(
		application,
		"configuration_header.md",
		"../../documentation/docs/configuration.md",
		ignoreFlags,
		replaceTopicNames,
	)

	createDefaultConfigFile(
		application,
		"../../config_defaults.json",
		ignoreFlags,
	)
}
