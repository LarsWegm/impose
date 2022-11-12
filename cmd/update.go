/*
Copyright © 2022 Lars Wegmann

*/
package cmd

import (
	"git.larswegmann.de/lars/impose/composeparser"
	"git.larswegmann.de/lars/impose/registry"
	"github.com/spf13/cobra"
)

var regCfg *registry.Config

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update image versions",
	Long:  `Updates the image versions in the specified Docker Compose file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parser, err := composeparser.NewParser(opts.InputFile)
		if err != nil {
			return err
		}
		r := registry.NewRegistry(regCfg)
		err = parser.UpdateVersions(r)
		if err != nil {
			return err
		}
		return writeOutput(parser)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	regCfg = &registry.Config{}
	updateCmd.Flags().StringVarP(&regCfg.Registry, "registry", "r", "https://hub.docker.com", "Docker registry to use for version lookup")
	updateCmd.Flags().StringVarP(&regCfg.User, "user", "u", "", "Docker registry user")
	updateCmd.Flags().StringVarP(&regCfg.Password, "password", "p", "", "Docker registry password")
}
