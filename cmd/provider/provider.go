package main

import (
	"github.com/ozankasikci/auth-provider-container/internal"
	"os"
)

func main() {
	os.Setenv("GOOGLE_CLIENT_ID", "865902519436-kia9b4p6a9vejck7ep656hheo4pdjlpd.apps.googleusercontent.com")
	os.Setenv("GOOGLE_CLIENT_SECRET", "Pv2uMlLwGAMgpnwEPN_tYWQs")
	internal.Start()
}
