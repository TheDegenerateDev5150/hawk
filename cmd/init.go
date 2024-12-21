/*
Copyright © 2023 Nick Godzieba <nick.godzieba@outlook.de>
*/
package cmd

import (
	"github.com/niiigoo/hawk/kit"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a project",
	Long: `Creates a new project and initializes it.
Adds go.mod with the given repository path as module name.
In addition, the basic proto file is added. If a name is provided, it is used, the the basename of the given repository otherwise.

Example:
hawk init github.com/my-org/test-service test
Output:
 - go.mod
   module: github.com/my-org/test-service
 - test.proto
   service: Test`,
	Run: func(cmd *cobra.Command, args []string) {
		g := kit.NewGenerator()
		err := g.Init(args...)
		printErrorAndExit(err)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
