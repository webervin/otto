package main

import (
	"os"
	"os/signal"

	appCustom "github.com/hashicorp/otto/builtin/app/custom"
	appDockerExt "github.com/hashicorp/otto/builtin/app/docker-external"
	appGo "github.com/hashicorp/otto/builtin/app/go"
	appNode "github.com/hashicorp/otto/builtin/app/node"
	appPHP "github.com/hashicorp/otto/builtin/app/php"
	appRuby "github.com/hashicorp/otto/builtin/app/ruby"
	appPython "github.com/hashicorp/otto/builtin/app/python"

	foundationConsul "github.com/hashicorp/otto/builtin/foundation/consul"
	infraAws "github.com/hashicorp/otto/builtin/infra/aws"

	"github.com/hashicorp/otto/builtin/pluginmap"
	"github.com/hashicorp/otto/command"
	"github.com/hashicorp/otto/foundation"
	"github.com/hashicorp/otto/infrastructure"
	"github.com/hashicorp/otto/otto"
	"github.com/mitchellh/cli"
)

// Commands is the mapping of all the available Otto commands.
var Commands map[string]cli.CommandFactory
var CommandsInclude []string
var Detectors = []*detect.Detector{
	&detect.Detector{
		Type: "go",
		File: []string{"*.go"},
	},
	&detect.Detector{
		Type: "php",
		File: []string{"*.php", "composer.json"},
	},
	&detect.Detector{
		Type: "ruby",
		File: []string{"*.rb", "Gemfile", "config.ru"},
	},
	&detect.Detector{
		Type: "node",
		File: []string{"package.json"},
	},
	&detect.Detector{
		Type: "python",
		File: []string{"*.py", "requirements.txt"},
	},
}

// Ui is the cli.Ui used for communicating to the outside world.
var Ui cli.Ui

const (
	ErrorPrefix  = "e:"
	OutputPrefix = "o:"
)

func init() {
	Ui = &cli.ColoredUi{
		OutputColor: cli.UiColorNone,
		InfoColor:   cli.UiColorNone,
		ErrorColor:  cli.UiColorRed,
		WarnColor:   cli.UiColorNone,
		Ui: &cli.PrefixedUi{
			AskPrefix:    OutputPrefix,
			OutputPrefix: OutputPrefix,
			InfoPrefix:   OutputPrefix,
			ErrorPrefix:  ErrorPrefix,
			Ui:           &cli.BasicUi{Writer: os.Stdout},
		},
	}

	apps := appGo.Tuples.Map(app.StructFactory(new(appGo.App)))
	apps.Add(appCustom.Tuples.Map(app.StructFactory(new(appCustom.App))))
	apps.Add(appDockerExt.Tuples.Map(app.StructFactory(new(appDockerExt.App))))
	apps.Add(appNode.Tuples.Map(app.StructFactory(new(appNode.App))))
	apps.Add(appPHP.Tuples.Map(app.StructFactory(new(appPHP.App))))
	apps.Add(appRuby.Tuples.Map(app.StructFactory(new(appRuby.App))))
	apps.Add(appPython.Tuples.Map(app.StructFactory(new(appPython.App))))

	foundations := foundationConsul.Tuples.Map(foundation.StructFactory(new(foundationConsul.Foundation)))

	meta := command.Meta{
		CoreConfig: &otto.CoreConfig{
			Foundations: foundations,
			Infrastructures: map[string]infrastructure.Factory{
				"aws": infraAws.Infra,
			},
		},
		Ui:        Ui,
		PluginMap: pluginmap.Map,
	}

	CommandsInclude = []string{
		"compile",
		"build",
		"deploy",
		"dev",
		"infra",
		"status",
		"version",
	}

	Commands = map[string]cli.CommandFactory{
		"compile": func() (cli.Command, error) {
			return &command.CompileCommand{
				Meta: meta,
			}, nil
		},

		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
				Meta: meta,
			}, nil
		},

		"deploy": func() (cli.Command, error) {
			return &command.DeployCommand{
				Meta: meta,
			}, nil
		},

		"dev": func() (cli.Command, error) {
			return &command.DevCommand{
				Meta: meta,
			}, nil
		},

		"infra": func() (cli.Command, error) {
			return &command.InfraCommand{
				Meta: meta,
			}, nil
		},

		"status": func() (cli.Command, error) {
			return &command.StatusCommand{
				Meta: meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:              meta,
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				CheckFunc:         commandVersionCheck,
			}, nil
		},

		// Internal or not shown to users directly

		"plugin-builtin": func() (cli.Command, error) {
			return &command.PluginBuiltinCommand{
				Meta: meta,
			}, nil
		},

		"help": func() (cli.Command, error) {
			return &command.HelpCommand{
				Meta: meta,
			}, nil
		},
	}
}

// makeShutdownCh creates an interrupt listener and returns a channel.
// A message will be sent on the channel for every interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
