package main

import (
	"encoding/json"
	"github.com/ozankasikci/one-oauth/internal/proxy"
	githubprovider "github.com/ozankasikci/one-oauth/provider/github"
	"github.com/ozankasikci/one-oauth/provider/google"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Open our jsonFile
	jsonFile, err := os.Open("cmd/proxy/creds.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var res struct {
		Google struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"google"`
		Github struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"github"`
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	providerConfig := proxy.NewConfig(
		"4999",
		proxy.AddGoogleConfig(&googleprovider.Config{
			ClientID:                   res.Google.ClientID,
			ClientSecret:               res.Google.ClientSecret,
			GoogleRedirectURL:          "http://localhost:5000/auth/google/callback",
			UpstreamSuccessRedirectURL: "http://localhost:5000/auth/google/success/callback",
			Scopes:                     []string{"profile", "email"},
			CookieSessionName:          "example-google-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "googleID",
		}),

		proxy.AddGithubConfig(&githubprovider.Config{
			ClientID:                   res.Github.ClientID,
			ClientSecret:               res.Github.ClientSecret,
			GithubRedirectURL:          "http://localhost:5000/auth/github/callback",
			UpstreamSuccessRedirectURL: "http://localhost:5000/auth/github/success/callback",
			Scopes: []string{"user"},
			CookieSessionName:          "example-github-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "githubID",
		}),
	)

	authProvider := proxy.New(providerConfig)
	authProvider.Start()
}
