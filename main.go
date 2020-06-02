package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 2 * 1024 * 1024 // 2 mb
const uploadPath = "/"

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

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			renderError(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
			return
		}

		// parse and validate file and post parameters
		file, header, err := r.FormFile("image")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		// Get and print out file size
		fileSize := header.Size
		fmt.Fprintf(w, "File size (bytes): %v\n", fileSize)
		// validate file size
		if fileSize > maxUploadSize {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}
		fileName := randToken(12)
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		workdir, _ := os.Getwd()
		newPath := filepath.Join(workdir, uploadPath, fileName+fileEndings[0])
		fmt.Fprintf(w, "FileType: %s, File: %s\n", detectedFileType, newPath)

		// write file
		newFile, err := os.Create(newPath) //illegal fix
		if err != nil {
			fmt.Fprintf(w, err.Error())
			renderError(w, "CANT_CREATE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("SUCCESS"))
		fmt.Fprintf(w, "File received. Args: size="+r.FormValue("size")+", iterations="+r.FormValue("iterations"))

	default:
		fmt.Fprintf(w, "Please POST your images")
	}
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
