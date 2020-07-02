package main

import (
	"github.com/ozankasikci/auth-provider-container/internal/provider"
	"github.com/ozankasikci/auth-provider-container/internal/provider/google"
)

func main() {
	providerConfig := provider.NewConfig(
		"4999",
		provider.AddGoogleConfig(&googleprovider.Config{
			ClientID:                   "",
			ClientSecret:               "",
			GoogleRedirectURL:          "http://localhost:5000/auth/google/callback",
			OneOauthSuccessRedirectURL: "http://localhost:5000/auth/google/success/callback",
			Scopes:                     []string{"profile", "email"},
			CookieSessionName:          "example-google-app",
			CookieSessionSecret:        "example cookie signing secret",
			CookieSessionUserKey:       "googleID",
		}),
	)

	authProvider := provider.New(providerConfig)
	authProvider.Start()
}
