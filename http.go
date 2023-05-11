package tools

import (
	"crypto/tls"
	"net/http"
	"os"
)

var httpClient *http.Client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: IsOn(os.Getenv("INSECURE"), false),
		},
	},
}

func HTTPClient() *http.Client {
	return httpClient
}
