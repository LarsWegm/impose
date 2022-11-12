/*
Copyright © 2022 Lars Wegmann

*/
package cmd

import (
	"git.larswegmann.de/lars/impose/composeparser"
	"github.com/spf13/cobra"
)

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Formats the Docker Compose file",
	Long: `Formats the Docker Compose file the same way an update would format it
without otherwise changing the content. This can be useful for taking a diff
(first format the file, then update the versions).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parser, err := composeparser.NewParser(opts.InputFile)
		if err != nil {
			return err
		}
		return writeOutput(parser)
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)
}
