package provider

import (
	"fmt"
	"github.com/gorilla/mux"
	googleprovider "github.com/ozankasikci/auth-provider-container/internal/provider/google"
	"log"
	"net/http"
)

type Config struct {
	Port         string
	GoogleConfig *googleprovider.Config
}

type Provider struct {
	Config         *Config
	Router         *mux.Router
	GoogleProvider googleprovider.GoogleProviderInterface
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

func New(config *Config) *Provider {
	router := mux.NewRouter()
	provider := &Provider{
		Config: config,
		Router: router,
	}

	if config.GoogleConfig != nil {
		googleProvider := googleprovider.New(config.GoogleConfig)
		router.Handle("/auth/google/login", googleProvider.LoginHandler())
		router.Handle("/auth/google/logout", googleProvider.LogoutHandler())
		router.Handle("/auth/google/callback", googleProvider.CallbackHandler())
		provider.GoogleProvider = googleProvider
	}

	return provider
}

func (t *Provider) Start() {
	address := fmt.Sprintf(":%s", t.Config.Port)

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, t.Router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
