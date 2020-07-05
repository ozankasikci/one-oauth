package main

import (
	"fmt"
	"github.com/dghubble/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	sessionName    = "example-google-app"
	sessionSecret  = "example cookie signing secret"
	sessionUserKey = "googleID"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

// New returns a new ServeMux with app routes.
func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", welcomeHandler)
	mux.Handle("/profile", requireLogin(http.HandlerFunc(profileHandler)))
	mux.HandleFunc("/logout", logoutHandler)

	app, _ := url.Parse("http://localhost:5000/")
	origin, _ := url.Parse("http://localhost:4999/")
	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", app.Host)
		req.Header.Add("X-Forwarded-For", app.Host)
		req.Header.Add("X-Origin-Host", app.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}
	proxy := &httputil.ReverseProxy{Director: director}

	mux.HandleFunc("/auth/google/login", func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) })
	mux.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) })
	mux.HandleFunc("/auth/google/success/callback", func(w http.ResponseWriter, r *http.Request) {
		println(r.URL.Query()["email"][0])

		http.Redirect(w, r, "/profile", http.StatusFound)
	})
	return mux
}

// welcomeHandler shows a welcome message and login button.
func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	if isAuthenticated(req) {
		http.Redirect(w, req, "/profile", http.StatusFound)
		return
	}
	page, _ := ioutil.ReadFile("home.html")
	fmt.Fprintf(w, string(page))
}

// profileHandler shows protected user content.
func profileHandler(w http.ResponseWriter, req *http.Request) {
	if !isAuthenticated(req) {
		fmt.Fprint(w, `You are not logged in :)`)
		return
	}
	fmt.Fprint(w, `<p>You are logged in!</p><form action="/logout" method="post"><input type="submit" value="Logout"></form>`)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

// requireLogin redirects unauthenticated users to the login route.
func requireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !isAuthenticated(req) {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func isAuthenticated(req *http.Request) bool {
	if _, err := sessionStore.Get(req, sessionName); err == nil {
		return true
	}
	return false
}

// main creates and starts a Server listening.
func main() {
	const address = "localhost:5000"
	// read credentials from environment variables if available

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New())
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
