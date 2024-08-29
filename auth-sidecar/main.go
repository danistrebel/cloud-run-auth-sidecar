package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

const (
	sidecarPort  = "8000" // Default port for the sidecar to listen on
	targetDomain = ".run.app"
)

func main() {
	http.HandleFunc("/", handleRequest)

	port := os.Getenv("AUTH_PROX_PORT")
	if port == "" {
		port = sidecarPort
	}

	log.Printf("Sidecar listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Check if the request is for the target domain
	if strings.HasSuffix(r.Host, targetDomain) {
		log.Println("Performing auth injection for calling Cloud Run Service:", r.Host)
		r.URL.Scheme = "https"
		// Get the GCP identity token
		token, err := getIdentityToken("https://" + r.Host)
		if err != nil {
			log.Printf("Error getting identity token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Authorization") == "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}

		// Proxy the modified request
		proxyRequest(w, r)
		return
	} else {
		log.Println("no auth injection for host:", r.Host)
		proxyRequest(w, r)
		return
	}

}

func getIdentityToken(aud string) (string, error) {

	ctx := context.Background()

	// Construct the GoogleCredentials object which obtains the default configuration from your
	// working environment.
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to generate default credentials: %w", err)
	}

	ts, err := idtoken.NewTokenSource(ctx, aud, option.WithCredentials(credentials))
	if err != nil {
		return "", fmt.Errorf("failed to create NewTokenSource: %w", err)
	}

	// Get the ID token.
	// Once you've obtained the ID token, you can use it to make an authenticated call
	// to the target audience.
	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.AccessToken, nil
}

func proxyRequest(w http.ResponseWriter, r *http.Request) {
	// Create a reverse proxy to forward the request
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: r.URL.Scheme,
		Host:   r.Host,
	})

	// Serve the proxied request
	proxy.ServeHTTP(w, r)
}
