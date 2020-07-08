package provider

import "net/http"

type ProviderInterface interface {
	LoginHandler() http.Handler
	LogoutHandler() http.Handler
	CallbackHandler() http.Handler
	IsAuthenticatedHandler() http.Handler
}

