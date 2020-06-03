package ingress

/*
INFO:	this module consists of code for receiving 2 images and 2 params via POST, and saving them locally
USAGE:	first call Setup() and then setup an http handler with StyleTransfer

		ingress.Setup()
		http.HandleFunc("/", ingress.StyleTransfer)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
*/

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
const uploadPath = "data"

/*Setup needs to be called to create a folder for storing received images*/
func Setup() {
	if err := createFolder(uploadPath); err != nil {
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
			renderError(w, "CANT_PARSE_FORM")
			return
		}
		//save files
		folder := randToken(12)
		folderPath := filepath.Join(uploadPath, folder)
		createFolder(folderPath)
		e1 := saveImage(w, r, "content", folder)
		e2 := saveImage(w, r, "style", folder)
		if e1 != nil && e2 != nil {
			renderError(w, "\nFAILED")
		} else {
			fmt.Fprintf(w, "Files received. Args: size="+r.FormValue("size")+", iterations="+r.FormValue("iterations"))
		}

	default:
		fmt.Fprintf(w, "Please POST your images")
	}
}

func saveImage(w http.ResponseWriter, r *http.Request, name string, folder string) error {
	file, header, err := r.FormFile(name)

	if err != nil {
		renderError(w, "INVALID_CONTENT_FILE")
		return err
	}
	defer file.Close()

	fileSize := header.Size
	fmt.Fprintf(w, "File: %s - size (bytes): %v\n", header.Filename, fileSize)
	if fileSize > maxUploadSize {
		renderError(w, "IMAGE_TOO_BIG")
		return err
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "INVALID_IMAGE")
		return err
	}
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/png":
	case "application/pdf":
		break
	default:
		renderError(w, "INVALID_FILE_TYPE")
		return err
	}
	fileName := name
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE")
		return err
	}
	workdir, _ := os.Getwd()
	newPath := filepath.Join(workdir, uploadPath, folder, fileName+fileEndings[0])
	fmt.Fprintf(w, "FileType: %s - Path to file: %s\n", detectedFileType, newPath)
	newFile, err := os.Create(newPath)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		renderError(w, "CANT_CREATE_FILE")
		return err
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(w, "CANT_WRITE_FILE")
		return err
	}
	return nil
}

func createFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0777) //777 = everyone can read/write/execute
	}
	//if path exists already that's fine
	return nil
}

func renderError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
