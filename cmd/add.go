package cmd

import (
	"fmt"
	"github.com/Bpazy/acr/unique"
	"github.com/Bpazy/acr/urls"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add clash rule to your rule-provider with file type",
	Long:  `Add clash rule to your rule-provider with file type`,
	Run:   addRule(),
}

var sortRules bool

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().BoolVarP(&sortRules, "sort", "", false, "Sort rules")
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

		// append rules to file
		appendRules(domains)

		// refresh clash core to enable new rules
		coreUrl := readClashCoreUrl()
		refreshRuleProvider(coreUrl)
	}
}

// append rules to file
func appendRules(domains []string) {
	p := getRuleProviderPath()

	b, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("Read rule file failed: %v", err)
	}
	var config map[string]*[]string
	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Fatalf("Unmarshal rule file failed: %v", err)
	}

	rules := config["payload"]
	for _, domain := range domains {
		*rules = append(*rules, "DOMAIN-SUFFIX,"+domain)
	}

	// distinct
	*rules = unique.Strings(*rules)
	if sortRules {
		sort.Strings(*rules)
	}

	b, err = yaml.Marshal(config)
	if err != nil {
		log.Fatalf("Marshal rules failed: %v", err)
	}
	if err := ioutil.WriteFile(p, b, 0600); err != nil {
		log.Fatalf("Write rules failed: %v", err)
	}
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
	name := "my-proxy"
	coreUrl := fmt.Sprintf("http://%s/providers/rules/%s", ec, name)
	return coreUrl
}

type CfwProfile struct {
	RuleProviders map[string]CfwRuleProvider `yaml:"rule-providers"`
}

type CfwRuleProvider struct {
	Type     string `yaml:"type"`
	Behavior string `yaml:"behavior"`
	Path     string `yaml:"path"`
}

type CfwList struct {
	Files []CfwListFile `yaml:"files"`
	Index int           `yaml:"index"`
}

type CfwListSelect struct {
	Name string `yaml:"name"`
	Now  string `yaml:"now"`
}

type CfwListFile struct {
	URL      string          `yaml:"url"`
	Time     string          `yaml:"time"`
	Name     string          `yaml:"name"`
	Selected []CfwListSelect `yaml:"selected"`
	Mode     string          `yaml:"mode"`
}

// Acquire rule provider(file) path
func getRuleProviderPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(filepath.Join(homeDir, "/.config/clash/profiles/list.yml"))
	if err != nil {
		log.Fatalf("Read Clash for Windows list.yml failed: %v", err)
	}
	c := CfwList{}
	if err := yaml.Unmarshal(b, &c); err != nil {
		log.Fatalf("Unmarshal Clash for Windows list.yml failed: %v", err)
	}
	if len(c.Files) <= c.Index {
		log.Fatalf("Clash for Windows list.yml is empty. Please check your CFW's profiles.")
	}

	file := c.Files[c.Index]
	b, err = ioutil.ReadFile(filepath.Join(homeDir, "/.config/clash/profiles/"+file.Time))
	if err != nil {
		log.Fatalf("Read selected CFW's profile %s failed: %v", file.Time, err)
	}

	var cfwProfile CfwProfile
	if err := yaml.Unmarshal(b, &cfwProfile); err != nil {
		log.Fatalf("Unmarshal selected CFW's profile %s failed: %v", file.Time, err)
	}
	// TODO
	r := cfwProfile.RuleProviders["my-proxy"]
	if filepath.IsAbs(r.Path) {
		return r.Path
	}

	return filepath.Join(homeDir, "/.config/clash", r.Path)
}
