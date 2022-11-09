/*
Copyright © 2022 Lars Wegmann

*/
package cmd

import (
	"fmt"

	"git.larswegmann.de/lars/impose/composeparser"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update image versions",
	Long:  `Updates the image versions in the specified Docker Compose file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parser, err := composeparser.NewParser(&composeparser.Config{
			ComposeFilePath: *cfg.FilePath,
		})
		if err != nil {
			return err
		}
		fmt.Println(parser.GetImageVersions())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
