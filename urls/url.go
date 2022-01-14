package urls

import (
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

func GetDomainSuffix(arg string) (string, error) {
	if !strings.HasPrefix(arg, "http://") && !strings.HasPrefix(arg, "https://") {
		return "", errors.New("Prefix 'https://' or : 'http://' is required")
	}

	u, err := url.Parse(arg)
	if err != nil {
		return "", errors.Wrap(err, "my msg")
	}

	hostname := u.Hostname()
	if hostname != "" {
		split := strings.Split(hostname, ".")
		domainSuffix := strings.Join(split[len(split)-2:], ".")
		return domainSuffix, nil
	}

	split := strings.Split(arg, ".")
	domainSuffix := strings.Join(split[len(split)-2:], ".")
	return domainSuffix, nil
}
