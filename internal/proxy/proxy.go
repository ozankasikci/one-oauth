package proxy

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ozankasikci/one-oauth/internal/provider"
	facebookprovider "github.com/ozankasikci/one-oauth/internal/provider/facebook"
	githubprovider "github.com/ozankasikci/one-oauth/internal/provider/github"
	googleprovider "github.com/ozankasikci/one-oauth/internal/provider/google"
	"log"
	"net/http"
)

type Config struct {
	UpstreamSuccessRedirectURL string
	Port                       string
	GoogleConfig               *googleprovider.Config
	GithubConfig               *githubprovider.Config
	FacebookConfig             *facebookprovider.Config
}

type Proxy struct {
	Config           *Config
	Router           *mux.Router
	GoogleProvider   provider.ProviderInterface
	GithubProvider   provider.ProviderInterface
	FacebookProvider provider.ProviderInterface
}

func NewConfig(port string, options ...func(*Config)) *Config {
	config := &Config{
		Port: port,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

func AddGoogleConfig(config *googleprovider.Config) func(*Config) {
	return func(c *Config) {
		c.GoogleConfig = config
	}
}

func AddGithubConfig(config *githubprovider.Config) func(*Config) {
	return func(c *Config) {
		c.GithubConfig = config
	}
}

func AddFacebookConfig(config *facebookprovider.Config) func(*Config) {
	return func(c *Config) {
		c.FacebookConfig = config
	}
}

func New(config *Config) *Proxy {
	router := mux.NewRouter()
	proxy := &Proxy{
		Config: config,
		Router: router,
	}

	if config.GoogleConfig != nil {
		googleProvider := googleprovider.New(config.GoogleConfig)
		router.Handle("/auth/google/login", googleProvider.LoginHandler())
		router.Handle("/auth/google/logout", googleProvider.LogoutHandler())
		router.Handle("/auth/google/callback", googleProvider.CallbackHandler())
		proxy.GoogleProvider = googleProvider
	}

	if config.GithubConfig != nil {
		githubProvider := githubprovider.New(config.GithubConfig)
		router.Handle("/auth/github/login", githubProvider.LoginHandler())
		router.Handle("/auth/github/logout", githubProvider.LogoutHandler())
		router.Handle("/auth/github/callback", githubProvider.CallbackHandler())
		proxy.GithubProvider = githubProvider
	}

	if config.FacebookConfig != nil {
		facebookProvider := facebookprovider.New(config.FacebookConfig)
		router.Handle("/auth/facebook/login", facebookProvider.LoginHandler())
		router.Handle("/auth/facebook/logout", facebookProvider.LogoutHandler())
		router.Handle("/auth/facebook/callback", facebookProvider.CallbackHandler())
		proxy.FacebookProvider = facebookProvider
	}

	return proxy
}

func (t *Proxy) Start() {
	address := fmt.Sprintf(":%s", t.Config.Port)

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, t.Router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
