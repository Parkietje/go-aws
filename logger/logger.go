package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

var file *os.File
var csvWriter *csv.Writer

func Init() {
	// Create the log csv file
	csvFile, err := os.Create("log.csv")
	if err != nil {
		fmt.Println(err)
	}

	// Save it as global variable
	file = csvFile
}

func Log(total string, active string) {
	csvWriter = csv.NewWriter(file)

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	row := []string{timestamp, total, active}

	_ = csvWriter.Write(row)

	csvWriter.Flush()
}

func Stop() {
	csvWriter.Flush()
	file.Close()
}
