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

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	folderPath := filepath.Join(userHomeDir(), "/.config/acr")
	viper.AddConfigPath(folderPath)
	if err := viper.ReadInConfig(); err != nil {
		viper.Set("proxy-name", "")

		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			log.Debugf("mkdir '%s' failed: %v", folderPath, err)
		}
		if err := viper.WriteConfigAs(filepath.Join(folderPath, "config.yaml")); err != nil {
			log.Fatalf("Create acrs's cofig file '~/.config/acr/config.yaml' failed: %v", err)
		}
		log.Debugf("Read acr's config failed: %v", err)
		log.Fatalf("'proxy-name' is empty. Please edit '~/.config/acr/config.yaml'")
	}
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

type CfwConfig struct {
	MixedPort          int    `yaml:"mixed-port"`
	AllowLan           bool   `yaml:"allow-lan"`
	ExternalController string `yaml:"external-controller"`
	Secret             string `yaml:"secret"`
	LogLevel           string `yaml:"log-level"`
}

// Acquire clash core API url
func readClashCoreUrl() string {
	p := filepath.Join(userHomeDir(), "/.config/clash/config.yaml")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("Read CFW's config file '%s' failed: %v", p, err)
	}
	config := CfwConfig{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Fatalf("Unmarshal CFW's config file '%s' failed: %v", p, err)
	}

	ec := config.ExternalController
	coreUrl := fmt.Sprintf("http://%s/providers/rules/%s", ec, proxyName())
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
	homeDir := userHomeDir()
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
	cfwProfilePath := filepath.Join(homeDir, "/.config/clash/profiles/"+file.Time)
	b, err = ioutil.ReadFile(cfwProfilePath)
	if err != nil {
		log.Fatalf("Read selected CFW's profile %s failed: %v", file.Time, err)
	}

	var cfwProfile CfwProfile
	if err := yaml.Unmarshal(b, &cfwProfile); err != nil {
		log.Fatalf("Unmarshal selected CFW's profile %s failed: %v", file.Time, err)
	}
	r, ok := cfwProfile.RuleProviders[proxyName()]
	if !ok {
		log.Fatalf("'%s' does not exist in %s", proxyName(), cfwProfilePath)
	}
	if filepath.IsAbs(r.Path) {
		return r.Path
	}

	return filepath.Join(homeDir, "/.config/clash", r.Path)
}

func userHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Acquire user home dir failed: %v", err)
	}
	return homeDir
}

func proxyName() string {
	p := viper.GetString("proxy-name")
	if p == "" {
		log.Fatalf("'proxy-name' is empty. Please edit ~/.config/acr/config.yaml")
	}
	return p
}
