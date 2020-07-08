package main

import (
	"encoding/json"
	facebookprovider "github.com/ozankasikci/one-oauth/internal/provider/facebook"
	githubprovider "github.com/ozankasikci/one-oauth/internal/provider/github"
	"github.com/ozankasikci/one-oauth/internal/provider/google"
	"github.com/ozankasikci/one-oauth/internal/proxy"
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
		Facebook struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"facebook"`
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
			Scopes:                     []string{"user"},
			CookieSessionName:          "example-github-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "githubID",
		}),

		proxy.AddFacebookConfig(&facebookprovider.Config{
			ClientID:                   res.Facebook.ClientID,
			ClientSecret:               res.Facebook.ClientSecret,
			FacebookRedirectURL:        "http://localhost:5000/auth/facebook/callback",
			UpstreamSuccessRedirectURL: "http://localhost:5000/auth/facebook/success/callback",
			Scopes:                     []string{"email"},
			CookieSessionName:          "example-facebook-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "facebookID",
		}),
	)

	authProvider := proxy.New(providerConfig)
	authProvider.Start()
}
