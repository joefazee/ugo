package main

func doMigrate(command string, step string) error {
	dsn := getDSN()
	switch command {
	case "up":
		return ug.MigrateUp(dsn)
	case "down":
		if step == "all" {
			return ug.MigrateDownAll(dsn)
		}
		return ug.MigrateStep(-1, dsn)
	case "reset":
		err := ug.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		return ug.MigrateUp(dsn)
	default:
		showHelp()
	}

	return nil
}

