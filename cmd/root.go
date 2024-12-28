/*
Copyright © 2023 Nick Godzieba nick.godzieba@outlook.de
*/
package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hawk",
	Short: "A go boilerplate generator based on a proto file",
	Long:  `Hawk is a CLI-tool to build microservices with go in a way, you can focus on the business logic. Multiple transport layers are supported, everything starts with a proto file! `,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hawk.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func printErrorAndExit(err error) {
	if err != nil {
		last := ""
		for ; err != nil; err = errors.Unwrap(err) {
			if !strings.Contains(last, err.Error()) {
				last = err.Error()
				println(last)
			}
		}
		os.Exit(1)
	}
}
