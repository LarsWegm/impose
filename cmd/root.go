/*
Copyright © 2022 Lars Wegmann

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "impose",
	Short: "Image version updater for Docker Compose",
	Long: `Image version updater for Docker Compose.
This tool automatically scans the given Docker Compose
file for image versions and updates them.

You can use head or inline comments for the image keyword in the Docker Compose file to add annotations.
The following annotations are available:
  impose:ignore     ignores the image for updates
  impose:minor      only checks for minor version updates
  impose:patch      only checks for patch version updates
  impose:warnMajor  warns if major version has changed
  impose:warnMinor  warns if minor version has changed (including major version changes)
  impose:warnPatch  warns if patch version has changed (including major and minor version changes)
  impose:warnAll    warns if the version string has changed in any way (including version suffix)`,
}

type CliOptions struct {
	InputFile  string
	OutputFile string
}

type writer interface {
	WriteToOriginalFile() error
	WriteToStdout() error
	WriteToFile(file string) error
}

var opts *CliOptions

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	opts = &CliOptions{}
	rootCmd.PersistentFlags().StringVarP(&opts.InputFile, "file", "f", "docker-compose.yml", "Compose file")
	rootCmd.PersistentFlags().StringVarP(&opts.OutputFile, "out", "o", "", "The output file (default is the input file, if \"-\" is passed it writes to std out)")
}

func writeOutput(w writer) (err error) {
	switch opts.OutputFile {
	case "":
		err = w.WriteToOriginalFile()
	case "-":
		err = w.WriteToStdout()
	default:
		err = w.WriteToFile(opts.OutputFile)
	}
	return
}
