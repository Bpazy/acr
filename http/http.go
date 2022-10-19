package http

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type Headers map[string]string

func Put(url string, h Headers) []byte {
	req := buildRequest(http.MethodPut, url, nil, h)
	return Do(req)
}

func Get(url string, h Headers) []byte {
	req := buildRequest(http.MethodGet, url, nil, h)
	return Do(req)
}

func Delete(url string, h Headers) []byte {
	req := buildRequest(http.MethodDelete, url, nil, h)
	return Do(req)
}

func Do(req *http.Request) []byte {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request %s failed: %v", req.URL, err)
	}
	defer res.Body.Close()
	log.Debugf("Do request to %s, response: %+v\n", req.URL, res)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Read %s's response failed: %v", req.URL, err)
	}

	return b
}

func buildRequest(method, url string, body io.Reader, headers Headers) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatalf("Build new request failed: %v", err)
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	return req
}
