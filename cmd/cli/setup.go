package main

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
)

func setup(arg1, arg2 string) {

	if arg1 != "new" && arg1 != "help" && arg1 != "version" {
		err := godotenv.Load()
		if err != nil {
			exitGracefully(errors.New("error loading .env file"))
		}

		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}

		ug.RootPath = path
		ug.DB.DataType = os.Getenv("DATABASE_TYPE")
	}

}
