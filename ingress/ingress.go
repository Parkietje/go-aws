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
	"go-aws/m/v2/loadbalancer"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	maxUploadSize = 2 * 1024 * 1024 // 2 mb
	uploadPath    = "data"
	style         = "style"
	content       = "content"
	combined      = "combined.png"
)

/*Setup needs to be called to create a folder for storing received images*/
func Setup() {
	if err := createFolder(uploadPath); err != nil {
		log.Fatal(err)
	}
}

/*StyleTransfer is a httphandler which accepts two images and responds with the result*/
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
		styleFile, _, _, e1 := saveImage(w, r, content, folder)
		contentFile, _, _, e2 := saveImage(w, r, style, folder)
		if e1 != nil || e2 != nil {
			renderError(w, "\nFAILED")
		} else {
			err := loadbalancer.RunApplication(folder, styleFile, contentFile)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			Openfile, err := os.Open(filepath.Join(folderPath, combined))
			defer Openfile.Close()
			if err != nil {
				http.Error(w, "File not found. @@", 404)
				return
			}
			//Get the Content-Type of the file
			//Create a buffer to store the header of the file in
			FileHeader := make([]byte, 512)
			//Copy the headers into the FileHeader buffer
			Openfile.Read(FileHeader)
			FileContentType := http.DetectContentType(FileHeader)
			FileStat, _ := Openfile.Stat()                     //Get info from file
			FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
			//Send the headers
			w.Header().Set("Content-Disposition", "attachment; filename="+combined)
			w.Header().Set("Content-Type", FileContentType)
			w.Header().Set("Content-Length", FileSize)
			//We read 512 bytes from the file already, so we reset the offset back to 0
			Openfile.Seek(0, 0)
			io.Copy(w, Openfile) //'Copy' the file to the client
			//os.RemoveAll(folderPath)
			return
		}

	default:
		fmt.Fprintf(w, "Please POST your images")
	}
}

func saveImage(w http.ResponseWriter, r *http.Request, name string, folder string) (fileName string, detectedFileType string, fileSize int64, err error) {
	file, header, err := r.FormFile(name)
	if err != nil {
		renderError(w, "INVALID_CONTENT_FILE")
		return
	}
	defer file.Close()
	fileSize = header.Size
	//fmt.Fprintf(w, "File: %s - size (bytes): %v\n", header.Filename, fileSize)
	if fileSize > maxUploadSize {
		renderError(w, "IMAGE_TOO_BIG")
		return
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(w, "INVALID_IMAGE")
		return
	}
	detectedFileType = http.DetectContentType(fileBytes)
	// TODO: shomehow this case statement gives problems when copying jpeg images, only jpg works for now
	switch detectedFileType {
	case "image/jpg", "image/jpeg":
	case "image/png":
	case "application/pdf":
		break
	default:
		renderError(w, "INVALID_FILE_TYPE")
		return
	}
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE")
		return
	}
	workdir, _ := os.Getwd()
	fileName = filepath.Join(workdir, uploadPath, folder, name+fileEndings[0])
	//fmt.Fprintf(w, "FileType: %s - Path to file: %s\n", detectedFileType, fileName)
	newFile, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		renderError(w, "CANT_CREATE_FILE")
		return
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(w, "CANT_WRITE_FILE")
	}
	fileName = name + fileEndings[0]
	return
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
