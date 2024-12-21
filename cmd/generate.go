/*
Copyright © 2023 Nick Godzieba <nick.godzieba@outlook.de>
*/
package cmd

import (
	"github.com/niiigoo/hawk/kit"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generates the go service based on the proto file",
	Long: `Generated files:

Project dir
├── cmd # do not touch, will be overridden
│   ├── <project>
│   │   ├── main.go
├── handlers
│   ├── handlers.go # entrypoint for business logic, endpoints are updated
│   ├── hooks.go # stop gracefully, generated only once
│   ├── middleware.go # apply middleware, generated only once
├── svc # do not touch, will be overridden
├── go.mod
├── go.sum
├── *.pb.go # do not touch, will be overridden`,
	Run: func(cmd *cobra.Command, args []string) {
		g := kit.NewGenerator()
		err := g.Service(args...)
		printErrorAndExit(err)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
