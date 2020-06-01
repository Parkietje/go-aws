package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", StyleTransfer)

	fmt.Println("starting server on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

/*StyleTransfer is a httphandler which accepts two images and sends them to the job queue*/
func StyleTransfer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found. ", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		fmt.Fprintf(w, "post succesful")
		//do processing of the POST here

	default:
		fmt.Fprintf(w, "Please POST your images")
	}
}
