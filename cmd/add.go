package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Bpazy/acr/http"
	"github.com/Bpazy/acr/unique"
	"github.com/Bpazy/acr/urls"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

	addCmd.Flags().BoolVarP(&sortRules, "sort", "", true, "Sort rules")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	folderPath := filepath.Join(userHomeDir(), "/.config/acr")
	viper.AddConfigPath(folderPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Debugf("Read acr's config failed: %v", err)
		viper.Set("proxy-name", "")

		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			log.Debugf("mkdir '%s' failed: %v", folderPath, err)
		}
		if err := viper.WriteConfigAs(filepath.Join(folderPath, "config.yaml")); err != nil {
			log.Fatalf("Create acrs's cofig file '~/.config/acr/config.yaml' failed: %v", err)
		}
		log.Fatalf("'proxy-name' is empty. Please edit '~/.config/acr/config.yaml'")
	}
}

func addRule() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		domains, ok := parseDomains(args)
		if !ok || len(domains) == 0 {
			return
		}
		// append rules to file
		appendRules(domains)
		// refresh clash core to enable new rules
		refreshRuleProvider(domains)
	}
}

func parseDomains(args []string) ([]string, bool) {
	log.Debugln("Add called: " + strings.Join(args, ", "))

	// Extract hostnames from parameters
	var domains []string
	for _, arg := range args {
		ds, err := urls.GetDomainSuffix(arg)
		if err != nil {
			log.Fatalf("%s: %s", err.Error(), arg)
			return nil, false
		}
		domains = append(domains, ds)
	}
	log.Debugf("Parsed hostnames: %v\n", domains)
	return domains, true
}

// append rules to file
func appendRules(domains []string) {
	p := getRuleProviderPath()

	b, err := os.ReadFile(p)
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
	if err := os.WriteFile(p, b, 0600); err != nil {
		log.Fatalf("Write rules failed: %v", err)
	}
}

func refreshRuleProvider(domains []string) {
	c := readCfwConfig()
	// Refresh rule providers
	log.Debugf("Refreshing rule provider: %s\n", c.providersRulesUrl())
	http.Put(c.providersRulesUrl(), c.headers())
	// Get all connections
	log.Debugf("Getting connections: %s\n", c.connectionsUrl())
	b := http.Get(c.connectionsUrl(), c.headers())
	ac := AllConnection{}
	if err := json.Unmarshal(b, &ac); err != nil {
		log.Fatalf("Get all CFW's connections failed: %v", err)
	}

	// Kill connections which added
	killConnections(&c, domains, ac.Connections)
}

func killConnections(c *CfwConfig, domains []string, connections []Connections) {
	for _, addedDomain := range domains {
		for _, connection := range connections {
			if strings.HasSuffix(connection.Metadata.Host, addedDomain) {
				log.Debugf("Killing connection: %s\n", c.connectionUrl(connection.ID))
				http.Delete(c.connectionUrl(connection.ID), c.headers())
			}
		}
	}
}

type AllConnection struct {
	DownloadTotal int           `json:"downloadTotal"`
	UploadTotal   int           `json:"uploadTotal"`
	Connections   []Connections `json:"connections"`
}

type Connections struct {
	ID          string    `json:"id"`
	Metadata    Metadata  `json:"metadata"`
	Upload      int       `json:"upload"`
	Download    int       `json:"download"`
	Start       time.Time `json:"start"`
	Chains      []string  `json:"chains"`
	Rule        string    `json:"rule"`
	RulePayload string    `json:"rulePayload"`
}

type Metadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	DNSMode         string `json:"dnsMode"`
}

type CfwConfig struct {
	MixedPort          int    `yaml:"mixed-port"`
	AllowLan           bool   `yaml:"allow-lan"`
	ExternalController string `yaml:"external-controller"`
	Secret             string `yaml:"secret"`
	LogLevel           string `yaml:"log-level"`
}

func (c CfwConfig) baseUrl() string {
	return "http://" + c.ExternalController
}

func (c CfwConfig) providersRulesUrl() string {
	return fmt.Sprintf("%s/providers/rules/%s", c.baseUrl(), proxyName())
}

func (c CfwConfig) connectionsUrl() string {
	return fmt.Sprintf("%s/connections", c.baseUrl())
}

func (c CfwConfig) connectionUrl(connectionId string) string {
	return fmt.Sprintf("%s/connections/%s", c.baseUrl(), connectionId)
}

func (c CfwConfig) headers() http.Headers {
	h := http.Headers{}
	if c.Secret != "" {
		h["Authorization"] = "Bearer " + c.Secret
	}
	return h
}

func readCfwConfig() CfwConfig {
	p := filepath.Join(userHomeDir(), "/.config/clash/config.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("Read CFW's config file '%s' failed: %v", p, err)
	}
	config := CfwConfig{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Fatalf("Unmarshal CFW's config file '%s' failed: %v", p, err)
	}
	return config
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
	b, err := os.ReadFile(filepath.Join(homeDir, "/.config/clash/profiles/list.yml"))
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
	b, err = os.ReadFile(cfwProfilePath)
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
