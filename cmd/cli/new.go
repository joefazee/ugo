package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"os"
	"os/exec"
	"strings"
)

var appUrl string

func doNew(appName string) error {

	appName = strings.ToLower(appName)
	appUrl = appName

	// sanitize the application name - convert url to single
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[len(exploded)-1]
	}

	// git clone skeleton application
	color.Green("cloning repo...")
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "https://github.com/joefazee/ugo-app.git",
		Progress: os.Stdout,
		Depth:    1,
	})

	if err != nil {
		return err
	}

	// remove .git directory
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		return err
	}

	// create a ready to go .env file
	color.Yellow("creating .env file...\n")
	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		return err
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", ug.GenerateRandomString(32))
	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		return err
	}
	// create a Makefile

	// update the go.mod file
	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		return err
	}
	mod := string(data)
	appUrl = cleanAppUrl(appUrl)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appUrl)

	err = copyDataToFile([]byte(mod), "./"+appName+"/go.mod")
	if err != nil {
		return err
	}

	// up existing .go files with correct name/imports
	color.Yellow("Updating source files...")
	err = os.Chdir("./" + appName)
	if err != nil {
		return err
	}
	err = updateSource()
	if err != nil {
		return err
	}

	// run go mod tidy in the project directory
	color.Yellow("Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		return err
	}

	color.Green("Done building " + appUrl)

	return nil
}
