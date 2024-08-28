package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		targetURL := os.Getenv("TARGET_URL")
		if targetURL == "" {
			fmt.Fprintln(w, "hello from the downstream app")
			return
		} else {
			log.Println("calling downstream app", targetURL)
		}
		resp, err := http.Get(targetURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("response from %s: ", targetURL)))
		w.Write(body)
	})

	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	fmt.Printf("Listening on port %s...", port)
	http.ListenAndServe(":"+port, nil)
}
