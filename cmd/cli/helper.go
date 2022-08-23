package main

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"
)

const (
	oldImport = "github.com/joefazee/ladiwork"
)

func getDSN() string {
	dbType := ug.DB.DataType

	if dbType == "pgx" || dbType == "postgres" {
		var dsn string
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		} else {
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		}

		return dsn
	}

	return "mysql://" + ug.BuildDSN()
}

func showHelp() {

	color.Yellow(`Available commands:

 help                   - show the help command
 version                - print application version 
 migrate | migrate up   - runs all up migrations that have not been run previously
 migrate down           - reverse the most recent migration
 migrate reset          - runs all down migrations in reverse order, and then all up migrations
 make migration <name>  - creates two new up and down migrations in the migrations folder
 make auth              - creates and runs migrations for authentication tables & create models and middleware
 make handler <name>	- creates a stub handler in the handlers directory
 make model <name>		- creates a new model in the data directory	
 make session 			- creates a table in the database as a session store
 make mail <name>		- create two starter mail templates in the mail directory
`)

}

func updateSourceFiles(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	matched, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	if matched {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := strings.Replace(string(data), oldImport, appUrl, -1)
		err = os.WriteFile(path, []byte(content), 0)
		if err != nil {
			return err
		}
	}

	return nil
}
func updateSource() error {
	return filepath.Walk(".", updateSourceFiles)
}

func cleanAppUrl(s string) string {
	if strings.Contains(s, "://") {
		s = strings.Replace(s, "https://", "", 1)
		s = strings.Replace(s, "http://", "", 1)
	}
	s = strings.TrimSuffix(s, "/")

	return s
}
