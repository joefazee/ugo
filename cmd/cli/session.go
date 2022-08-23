package main

import (
	"fmt"
	"strings"
	"time"
)

func doSessionTable() error {

	dbType := strings.ToLower(ug.DB.DataType)

	if dbType == "mariadb" {
		dbType = "mysql"
	}

	if dbType == "postgresql" {
		dbType = "postgres"
	}

	fileName := fmt.Sprintf("%d_create_sessions_table", time.Now().UnixMicro())

	upFile := ug.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
	downFile := ug.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	err := copyFileFromTemplate("templates/migrations/"+dbType+"_session.sql", upFile)
	if err != nil {
		return err
	}

	err = copyDataToFile([]byte("drop table sessions"), downFile)
	if err != nil {
		return err
	}

	err = doMigrate("up", "")

	return err
}
