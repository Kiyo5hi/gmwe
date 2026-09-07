package main

import (
	"encoding/json"
	"fmt"
	"gmwe/api/db"
	"gmwe/api/storage"
	"os"
)

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: storage import|check|backup <database-file> [backup-file]")
	}
	operation, path := os.Args[1], os.Args[2]
	if operation == "import" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		file.Close()
	}
	engine, err := db.Open("file:" + path + "?mode=rw")
	if err != nil {
		return err
	}
	pool, _ := engine.DB()
	defer pool.Close()
	switch operation {
	case "import":
		var data storage.Snapshot
		if err = json.NewDecoder(os.Stdin).Decode(&data); err != nil {
			return err
		}
		if err = storage.Import(engine, data); err != nil {
			return err
		}
	case "backup":
		if len(os.Args) != 4 {
			return fmt.Errorf("backup destination required")
		}
		if err = storage.Backup(engine, os.Args[3]); err != nil {
			return err
		}
	case "check":
	default:
		return fmt.Errorf("unknown operation")
	}
	if err = storage.Check(engine); err != nil {
		return err
	}
	snapshot, err := storage.Read(engine)
	if err != nil {
		return err
	}
	var version string
	engine.Raw("SELECT sqlite_version()").Scan(&version)
	return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"users": len(snapshot.Users), "hitokotos": len(snapshot.Hitokotos), "sha256": storage.Digest(snapshot), "sqlite": version})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "storage operation failed:", err)
		os.Exit(1)
	}
}
