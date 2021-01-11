package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of qingcloud-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("qingcloud-cli version is 0.0.0.1")
	},
}

