/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/Bpazy/acr/urls"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add clash rule to your rule-provider with file type",
	Long:  `Add clash rule to your rule-provider with file type`,
	Run:   addRule(),
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addRule() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		log.Debugln("Add called: " + strings.Join(args, ", "))

		// Extract hostnames from parameters
		var domains []string
		for _, arg := range args {
			ds, err := urls.GetDomainSuffix(arg)
			if err != nil {
				log.Fatalf("%s: %s", err.Error(), arg)
				return
			}
			domains = append(domains, ds)
		}
		log.Debugf("Parsed hostnames: %v\n", domains)

		if len(domains) == 0 {
			return
		}

		p := getRuleProviderPath()
		f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("Open rule file failed: %v", err)
		}
		defer f.Close()
		for _, domain := range domains {
			if _, err = f.WriteString("\n  - DOMAIN-SUFFIX," + domain); err != nil {
				log.Fatalf("Write rule failed: %v", err)
			}
		}

		coreUrl := readClashCoreUrl()
		refreshRuleProvider(coreUrl)
	}
}

// Acquire rule provider(file) path
func getRuleProviderPath() string {
	return "C:\\Users\\hanzi\\.config\\clash\\ruleset\\myproxy.yaml"
}

func refreshRuleProvider(coreUrl string) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, coreUrl, nil)
	if err != nil {
		log.Fatalf("Build new request failed: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request %s failed: %v", coreUrl, err)
	}
	defer res.Body.Close()

	log.Debugf("PUT %s response: %+v\n", coreUrl, res)
}

// Acquire clash core API url
func readClashCoreUrl() string {
	main := viper.New()
	main.SetConfigName("config")
	main.SetConfigType("yaml")
	main.AddConfigPath(filepath.FromSlash("$HOME/.config/clash"))
	if err := main.ReadInConfig(); err != nil {
		log.Fatalf("Read clash for Windows main config failed: %v", err)
	}

	ec := main.GetString("external-controller")
	coreUrl := fmt.Sprintf("http://%s/providers/rules/myproxy", ec)
	return coreUrl
}
