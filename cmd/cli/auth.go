package main

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

func doAuth() error {
	// migrations
	dbType := ug.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())

	upFile := ug.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := ug.RootPath + "/migrations/" + fileName + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/auth_tables."+dbType+".sql", upFile)
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/migrations/auth_tables_drop."+dbType+".sql", downFile)
	if err != nil {
		return err
	}

	err = doMigrate("up", "")
	if err != nil {
		return err
	}
	// run migrations

	err = copyFileFromTemplate("templates/data/user.go.txt", ug.RootPath+"/data/user.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/data/token.go.txt", ug.RootPath+"/data/token.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/data/remember_token.go.txt", ug.RootPath+"/data/remember_token.go")
	if err != nil {
		return err
	}

	// copy over middleware
	err = copyFileFromTemplate("templates/middleware/auth.go.txt", ug.RootPath+"/middleware/auth.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/middleware/remember.go.txt", ug.RootPath+"/middleware/remember.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/middleware/auth-token.go.txt", ug.RootPath+"/middleware/auth-token.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/handlers/auth-handlers.go.txt", ug.RootPath+"/handlers/auth-handlers.go")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.html.tmpl", ug.RootPath+"/mail/password-reset.html.tmpl")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/mailer/password-reset.plain.txt", ug.RootPath+"/mail/password-reset.plain.txt")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/views/forgot.page.jet", ug.RootPath+"/views/forgot.page.jet")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/views/reset-password.page.jet", ug.RootPath+"/views/reset-password.page.jet")
	if err != nil {
		return err
	}

	err = copyFileFromTemplate("templates/views/login.page.jet", ug.RootPath+"/views/login.page.jet")
	if err != nil {
		return err
	}

	color.Yellow(" - users, tokens, and remember_tokens migrations created and executed")
	color.Yellow(" - user and token models created")
	color.Yellow(" - auth middleware created")
	color.Yellow("")
	color.Yellow("Don`t forget to add user and token models in data/models.go, and to add all needed middlewares to your routes")

	return nil

}
