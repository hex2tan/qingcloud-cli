package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var echoDemoCmd = &cobra.Command{
	Use:   "echo-demo-config",
	Short: "echo demo configuration to standard output",
	Long: "qingcloud-cli echo-demo-config > $HOME/.qingcloud.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("qy_access_key_id: 'QYACCESSKEYIDEXAMPLE'\nqy_secret_access_key: 'SECRETACCESSKEY'\nzone: 'pek3'\n\n")
	},
}