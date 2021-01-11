package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	// Used for flags.
	cfgFile     string
	zone        string
	testCfgFile string

	rootCmd = &cobra.Command{
		Use:       "qingcloud-cli",
		Long:      "qingcloud-cli is a cli utility, you can run, describe, terminate instance",
		ValidArgs: []string{"run-instances"},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.qingcloud-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&zone, "zone", "", "specified the zone, overwrite the config file value")

	flagName := "zone"
	rootCmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validZoneList, cobra.ShellCompDirectiveDefault
	})

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(echoDemoCmd)
	addInstanceCmd(rootCmd)
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		fmt.Println("Use config file from the flag.")
		viper.SetConfigFile(cfgFile)
	} else if testCfgFile != "" {
		fmt.Println("Use config file from the test flag.")
		viper.SetConfigFile(testCfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".qingcloud")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed(), err)
		//fmt.Println("Must specified config file with --config flag or create a .qingcloud.yaml in $HOME directory.")
		//os.Exit(1)
	}
}
