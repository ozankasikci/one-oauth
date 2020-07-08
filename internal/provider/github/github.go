package githubprovider

import (
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/dghubble/sessions"
	"github.com/ozankasikci/one-oauth/internal/provider"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	"net/http"
	"net/url"
	"strconv"
)

type Config struct {
	CookieSessionName          string
	CookieSessionSecret        string
	CookieSessionUserKey       string
	ClientID                   string
	ClientSecret               string
	GithubRedirectURL          string
	UpstreamSuccessRedirectURL string
	Scopes                     []string
}

type GithubProvider struct {
	Config       *Config
	StateConfig  gologin.CookieConfig
	Oauth2Config *oauth2.Config
	CookieStore  *sessions.CookieStore
}

func (t GithubProvider) LoginHandler() http.Handler {
	return github.StateHandler(t.StateConfig, github.LoginHandler(t.Oauth2Config, nil))
}

func (t GithubProvider) LogoutHandler() http.Handler {
	return t.logoutHandler()
}

func (t GithubProvider) CallbackHandler() http.Handler {
	return github.StateHandler(t.StateConfig, github.CallbackHandler(t.Oauth2Config, t.issueSession(), nil))
}

func (t GithubProvider) IsAuthenticatedHandler() http.Handler {
	return t.IsAuthenticatedHandler()
}

func New(config *Config) provider.ProviderInterface {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.GithubRedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       config.Scopes,
	}

	cookieStore := sessions.NewCookieStore([]byte(config.CookieSessionSecret), nil)

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig

	return GithubProvider{
		Config:       config,
		StateConfig:  stateConfig,
		Oauth2Config: oauth2Config,
		CookieStore:  cookieStore,
	}
}

// issueSession issues a cookie session after successful github login
func (t *GithubProvider) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := t.CookieStore.New(t.Config.CookieSessionName)
		cookie.Values[t.Config.CookieSessionUserKey] = githubUser.ID
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
		q.Set("email", githubUser.GetEmail())
		q.Set("name", githubUser.GetName())
		q.Set("id", strconv.FormatInt(githubUser.GetID(), 10))
		q.Set("picture", githubUser.GetAvatarURL())
		q.Set("company", githubUser.GetCompany())
		q.Set("location", githubUser.GetLocation())
		q.Set("bio", githubUser.GetBio())
		q.Set("bio", githubUser.GetURL())
		successRedirectUrl.RawQuery = q.Encode()

		http.Redirect(w, r, successRedirectUrl.String(), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func (t *GithubProvider) logoutHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			t.CookieStore.Destroy(w, t.Config.CookieSessionName)
		}
	}

	return http.HandlerFunc(fn)
}

func (t *GithubProvider) isAuthenticatedHandler() http.Handler {
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
func (t *GithubProvider) isAuthenticated(r *http.Request) bool {
	if _, err := t.CookieStore.Get(r, t.Config.CookieSessionName); err == nil {
		return true
	}
	return false
}
