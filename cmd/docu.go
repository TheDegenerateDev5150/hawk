/*
Copyright © 2023 Nick Godzieba <nick.godzieba@outlook.de>
*/
package cmd

import (
	"github.com/niiigoo/hawk/docu"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var documentationCmd = &cobra.Command{
	Use:     "documentation",
	Aliases: []string{"docu", "d"},
	Short:   "Generates the documentation",
	Long: `Generates the documentation based on the comments in the proto file.
The format can be selected with the flag --format (or -f). Markdown is used by default.
Supported values are
 - md (template)
 - html

Examples:
hawk docu
hawk docu -f md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return err
		}
		g := docu.NewService()
		err = g.Generate(format, version, args...)
		if err != nil && errors.Is(err, docu.ErrFormat) {
			return err
		}
		printErrorAndExit(err)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(documentationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// documentationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// documentationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	documentationCmd.Flags().StringP("format", "f", "md", "Specify the output format. Supported values are md (MarkDown) and html")
	documentationCmd.Flags().StringP("version", "v", "", "Add a version number to the documentation")
}
