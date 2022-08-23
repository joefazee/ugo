package main

import (
	"embed"
	"errors"
	"io/ioutil"
	"os"
)

//go:embed templates
var templateFS embed.FS


func copyFileFromTemplate(fromFile string, toFile string) error {
	if fileExists(toFile) {
		return errors.New(toFile + " already exists")
	}

	data, err := templateFS.ReadFile(fromFile)
	if err != nil {
		return err
	}

	err = copyDataToFile(data, toFile)
	if err != nil {
		return err
	}
	return nil
}

func copyDataToFile(data []byte, file string) error {
	err := ioutil.WriteFile(file, data, 0644 )

	if err != nil {
		return err
	}

	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

