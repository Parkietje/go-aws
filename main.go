package main

import (
	"go-aws/m/v2/ingress"

	"log"
	"net/http"
)

func main() {

	// Start listening for post requests
	ingress.Setup()
	http.HandleFunc("/", ingress.StyleTransfer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
