package main

import (
	"encoding/json"
	"github.com/ozankasikci/one-oauth/internal/proxy"
	"github.com/ozankasikci/one-oauth/internal/proxy/google"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Open our jsonFile
	jsonFile, err := os.Open("cmd/proxy/google-creds.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var res struct {
		Web struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"web"`
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	providerConfig := proxy.NewConfig(
		"4999",
		proxy.AddGoogleConfig(&googleprovider.Config{
			ClientID:                   res.Web.ClientID,
			ClientSecret:               res.Web.ClientSecret,
			GoogleRedirectURL:          "http://localhost:5000/auth/google/callback",
			UpstreamSuccessRedirectURL: "http://localhost:5000/auth/google/success/callback",
			Scopes:                     []string{"profile", "email"},
			CookieSessionName:          "example-google-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "googleID",
		}),
	)

	authProvider := proxy.New(providerConfig)
	authProvider.Start()
}
