package googleprovider

import (
	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/google"
	"github.com/dghubble/sessions"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
	"net/http"
	"net/url"
)

type GoogleProviderInterface interface {
	LoginHandler() http.Handler
	LogoutHandler() http.Handler
	CallbackHandler() http.Handler
}

type Config struct {
	CookieSessionName    string
	CookieSessionSecret  string
	CookieSessionUserKey string
	ClientID             string
	ClientSecret         string
	GoogleRedirectURL    string
	SuccessRedirectURL   string
	Scopes               []string
}

type GoogleProvider struct {
	Config       *Config
	StateConfig  gologin.CookieConfig
	Oauth2Config *oauth2.Config
	CookieStore  *sessions.CookieStore
}

func (t GoogleProvider) LoginHandler() http.Handler {
	return google.StateHandler(t.StateConfig, google.LoginHandler(t.Oauth2Config, nil))
}

func (t GoogleProvider) LogoutHandler() http.Handler {
	return t.logoutHandler()
}

func (t GoogleProvider) CallbackHandler() http.Handler {
	return google.StateHandler(t.StateConfig, google.CallbackHandler(t.Oauth2Config, t.issueSession(), nil))
}

func New(config *Config) GoogleProviderInterface {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.GoogleRedirectURL,
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       config.Scopes,
	}

	cookieStore := sessions.NewCookieStore([]byte(config.CookieSessionSecret), nil)

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig

	return GoogleProvider{
		Config:       config,
		StateConfig:  stateConfig,
		Oauth2Config: oauth2Config,
		CookieStore:  cookieStore,
	}
}

// issueSession issues a cookie session after successful Google login
func (t *GoogleProvider) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		googleUser, err := google.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := t.CookieStore.New(t.Config.CookieSessionName)
		cookie.Values[t.Config.CookieSessionUserKey] = googleUser.Id
		err = cookie.Save(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		successRedirectUrl, err := url.Parse(t.Config.SuccessRedirectURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		verifiedEmailString := "false"
		if googleUser.VerifiedEmail != nil && *googleUser.VerifiedEmail == true {
			verifiedEmailString = "true"
		}

		q := successRedirectUrl.Query()
		q.Set("email", googleUser.Email)
		q.Set("name", googleUser.Name)
		q.Set("family_name", googleUser.FamilyName)
		q.Set("gender", googleUser.Gender)
		q.Set("given_name", googleUser.GivenName)
		q.Set("hd", googleUser.Hd)
		q.Set("id", googleUser.Id)
		q.Set("link", googleUser.Link)
		q.Set("locale", googleUser.Locale)
		q.Set("picture", googleUser.Picture)
		q.Set("verified_email", verifiedEmailString)
		successRedirectUrl.RawQuery = q.Encode()

		http.Redirect(w, r, successRedirectUrl.String(), http.StatusFound)
	}

	return http.HandlerFunc(fn)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func (t *GoogleProvider) logoutHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			t.CookieStore.Destroy(w, t.Config.CookieSessionName)
		}
	}

	return http.HandlerFunc(fn)
}

func (t *GoogleProvider) isAuthenticatedHandler() http.Handler {
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
func (t *GoogleProvider) isAuthenticated(r *http.Request) bool {
	if _, err := t.CookieStore.Get(r, t.Config.CookieSessionName); err == nil {
		return true
	}
	return false
}
