package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func getCurrentPath(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host

}

func zipAnswerWriter(storePath string) (string, error) {
	var files []string

	path := storePath + "/" + answerDir
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}

	zipFileName := "answers.zip"
	return ZipFiles(storePath, zipFileName, files)

}
func ZipFiles(storeDir string, filename string, files []string) (string, error) {
	fileDir := storeDir + "/temp"
	filePath := fileDir + "/" + filename
	os.MkdirAll(fileDir, os.ModePerm)
	newZipFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return "", err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return "", err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return "", err
		}

		header.Name = info.Name()

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return "", err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return "", err
		}
	}
	return filePath, nil
}
