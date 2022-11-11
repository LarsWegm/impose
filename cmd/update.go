/*
Copyright © 2022 Lars Wegmann

*/
package cmd

import (
	"git.larswegmann.de/lars/impose/composeparser"
	"git.larswegmann.de/lars/impose/registry"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update image versions",
	Long:  `Updates the image versions in the specified Docker Compose file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := registry.NewRegistry(registry.Config{})
		parser, err := composeparser.NewParser(*cfg.FilePath, r)
		if err != nil {
			return err
		}
		err = parser.UpdateVersions()
		if err != nil {
			return err
		}
		parser.WriteToStdout()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
