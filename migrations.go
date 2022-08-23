package ugo

import (
	"github.com/golang-migrate/migrate/v4"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)
func (u *Ugo) MigrateUp(dsn string) error {
	m, err := migrate.New("file://" + u.RootPath + "/migrations", dsn)
	if err != nil {
		return err
	}

	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration:", err)
		return err
	}

	return nil
}

func (u *Ugo) MigrateDownAll(dsn string) error  {

	m, err := migrate.New("file://" + u.RootPath + "/migrations", dsn)
	if err != nil {
		return err
	}

	defer m.Close()

	if err := m.Down(); err != nil {
		log.Println("Error running down migration:", err)
		return err
	}

	return nil

}


func (u *Ugo) MigrateStep(n int, dsn string) error  {

	m, err := migrate.New("file://" + u.RootPath + "/migrations", dsn)
	if err != nil {
		return err
	}

	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil

}

func (u *Ugo) MigrateForce(dsn string) error {

	m, err := migrate.New("file://" + u.RootPath + "/migrations", dsn)
	if err != nil {
		return err
	}

	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}
