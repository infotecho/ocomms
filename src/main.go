package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	server := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte("Hello world 2"))
			if err != nil {
				log.Printf("ResponseWriter failed: %v", err)
			}
		}),
		ReadHeaderTimeout: 1 * time.Second,
	}
	err := server.ListenAndServe()
	log.Fatal(err)
}
