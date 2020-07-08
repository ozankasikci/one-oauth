package facebookprovider

import (
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/facebook"
	"github.com/dghubble/sessions"
	"github.com/ozankasikci/one-oauth/internal/provider"
	"golang.org/x/oauth2"
	facebookOAuth2 "golang.org/x/oauth2/facebook"
	"net/http"
	"net/url"
)

type Config struct {
	CookieSessionName          string
	CookieSessionSecret        string
	CookieSessionUserKey       string
	ClientID                   string
	ClientSecret               string
	FacebookRedirectURL        string
	UpstreamSuccessRedirectURL string
	Scopes                     []string
}

type FacebookProvider struct {
	Config       *Config
	StateConfig  gologin.CookieConfig
	Oauth2Config *oauth2.Config
	CookieStore  *sessions.CookieStore
}

func (t FacebookProvider) LoginHandler() http.Handler {
	return facebook.StateHandler(t.StateConfig, facebook.LoginHandler(t.Oauth2Config, nil))
}

func (t FacebookProvider) LogoutHandler() http.Handler {
	return t.logoutHandler()
}

func (t FacebookProvider) CallbackHandler() http.Handler {
	return facebook.StateHandler(t.StateConfig, facebook.CallbackHandler(t.Oauth2Config, t.issueSession(), nil))
}

func (t FacebookProvider) IsAuthenticatedHandler() http.Handler {
	return t.IsAuthenticatedHandler()
}

func New(config *Config) provider.ProviderInterface {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.FacebookRedirectURL,
		Endpoint:     facebookOAuth2.Endpoint,
		Scopes:       config.Scopes,
	}

	cookieStore := sessions.NewCookieStore([]byte(config.CookieSessionSecret), nil)

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig

	return FacebookProvider{
		Config:       config,
		StateConfig:  stateConfig,
		Oauth2Config: oauth2Config,
		CookieStore:  cookieStore,
	}
}

// issueSession issues a cookie session after successful facebook login
func (t *FacebookProvider) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		facebookUser, err := facebook.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := t.CookieStore.New(t.Config.CookieSessionName)
		cookie.Values[t.Config.CookieSessionUserKey] = facebookUser.ID
		err = cookie.Save(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		successRedirectUrl, err := url.Parse(t.Config.UpstreamSuccessRedirectURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q := successRedirectUrl.Query()
		q.Set("email", facebookUser.Email)
		q.Set("name", facebookUser.Name)
		q.Set("id", facebookUser.ID)
		successRedirectUrl.RawQuery = q.Encode()

		http.Redirect(w, r, successRedirectUrl.String(), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func (t *FacebookProvider) logoutHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			t.CookieStore.Destroy(w, t.Config.CookieSessionName)
		}
	}

	return http.HandlerFunc(fn)
}

func (t *FacebookProvider) isAuthenticatedHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if t.isAuthenticated(r) == true {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusProxyAuthRequired)
		}
	}

	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func (t *FacebookProvider) isAuthenticated(r *http.Request) bool {
	if _, err := t.CookieStore.Get(r, t.Config.CookieSessionName); err == nil {
		return true
	}
	return false
}
