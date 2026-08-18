package main

import (
	"fmt"
	"os"

	"github.com/harvester/addons/pkg/render"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	templateSource = "./addons"
)

func main() {
	var generateAddons, generateTemplates, validate bool
	var path string
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "generateAddons",
				Value:       false,
				Usage:       "generate disabled addon yaml manifests",
				Destination: &generateAddons,
			},
			&cli.BoolFlag{
				Name:        "generateTemplates",
				Value:       false,
				Usage:       "generate template manifests",
				Destination: &generateTemplates,
			},
			&cli.BoolFlag{
				Name:        "validate",
				Value:       false,
				Usage:       "validate the addon directory layout and metadata",
				Destination: &validate,
			},
			&cli.StringFlag{
				Name:        "path",
				Value:       ".",
				Usage:       "destination for output files",
				Destination: &path,
			},
		},

		Action: func(ctx *cli.Context) error {
			if !generateAddons && !generateTemplates && !validate {
				return fmt.Errorf("generateAddons, generateTemplates or validate need to be specified")
			}

			// generation always validates the layout first, so invalid
			// metadata can never produce ISO or upgrade artifacts
			if err := render.Validate(templateSource, "version_info"); err != nil {
				return fmt.Errorf("error validating addons: %v", err)
			}

			if generateAddons {
				if err := render.Addon(templateSource, path, "version_info"); err != nil {
					return fmt.Errorf("error during rendering addons: %v", err)
				}
			}

			if generateTemplates {
				if err := render.Template(templateSource, path, "version_info"); err != nil {
					return fmt.Errorf("error during rendering template: %v", err)
				}
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

}
