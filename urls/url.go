package urls

import (
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

func GetDomainSuffix(arg string) (string, error) {
	u, err := url.ParseRequestURI(arg)
	if err != nil {
		log.Debugf("Skip parse uri beause no scahema found")
		return arg, nil
	}

	hostname := u.Hostname()
	if hostname != "" {
		split := strings.Split(hostname, ".")
		domainSuffix := strings.Join(split[len(split)-2:], ".")
		if domainSuffix == "com.cn" {
			return strings.Join(split[len(split)-3:], "."), nil
		}
		return domainSuffix, nil
	}

	split := strings.Split(arg, ".")
	domainSuffix := strings.Join(split[len(split)-2:], ".")
	return domainSuffix, nil
}
