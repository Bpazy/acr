/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/bpazy/acr/urls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: addRule(),
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func addRule() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		fmt.Println("add called: " + strings.Join(args, ", "))

		// Extract hostnames from parameters
		var domains []string
		for _, arg := range args {
			ds, err := urls.GetDomainSuffix(arg)
			if err != nil {
				fmt.Printf("%s: %s \n", err.Error(), arg)
				return
			}
			domains = append(domains, ds)
		}
		fmt.Printf("Parsed hostnames: %v\n", domains)

		// Read Clash for Windows config
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(filepath.FromSlash("$HOME/.config/clash"))
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Read clash for Windows config failed: %v", err)
		}

		ec := viper.GetString("external-controller")
		fmt.Println(ec)
	}
}
