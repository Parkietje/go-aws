package main

import (
	"fmt"
	"go-aws/m/v2/ingress"
	"log"
	"net/http"
)

func main() {

	ingress.Setup()
	http.HandleFunc("/", ingress.StyleTransfer)

	fmt.Println("starting server on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
