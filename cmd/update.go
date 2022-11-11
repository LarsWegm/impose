/*
Copyright © 2022 Lars Wegmann

*/
package cmd

import (
	"fmt"

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
		parser, err := composeparser.NewParser(*cfg.FilePath)
		if err != nil {
			return err
		}
		_, err = parser.GetImageVersions()
		if err != nil {
			return err
		}
		r := registry.NewRegistry(registry.Config{})
		img, err := registry.NewImageFromString("library/mariadb:10.5.13-jammy")
		if err != nil {
			return err
		}
		tags, err := r.GetLatestVersion(img)
		if err != nil {
			return err
		}
		fmt.Println(tags)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
