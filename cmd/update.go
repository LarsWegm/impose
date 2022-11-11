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
		out, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}
		switch out {
		case "":
			err = parser.WriteToOriginalFile()
		case "-":
			err = parser.WriteToStdout()
		default:
			err = parser.WriteToFile(out)
		}
		return err
	},
}

func init() {
	updateCmd.Flags().StringP("out", "o", "", "The output file (default is the input file, if \"-\" is passed it writes to std out)")
	rootCmd.AddCommand(updateCmd)
}
