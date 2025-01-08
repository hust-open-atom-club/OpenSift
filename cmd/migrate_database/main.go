package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

func main() {
	flagConfigPath := flag.String("config", "", "config file path")
	flag.Parse()

	if *flagConfigPath == "" {
		fmt.Println("Please provide config file path, use -h for help")
		return
	}

	appDb, err := storage.NewAppDatabase(*flagConfigPath)

	if err != nil {
		panic(err)
	}

	result, err := os.ReadDir("migrations")
	if err != nil {
		fmt.Printf("Failed to read migrations directory: %v\n", err)
		fmt.Println("Please make sure you are running this command in the root directory of the project")
		panic(err)
	}

	type MigrationItem struct {
		Version  string
		Name     string
		FileName string
	}

	migrations := make([]MigrationItem, 0)

	for _, file := range result {
		if !file.IsDir() {
			continue
		}

		filePath := fmt.Sprintf("migrations/%s/migration.sql", file.Name())

		stat, err := os.Stat(filePath)
		if err != nil || stat.IsDir() {
			continue
		}

		reg := regexp.MustCompile(`^(\d{4}_\d{2}_\d{2}_\d{2})_(.+)$`)
		matches := reg.FindStringSubmatch(file.Name())

		if len(matches) != 3 {
			fmt.Errorf("Invalid migration file name: %s", file.Name())
			continue
		}

		migrations = append(migrations, MigrationItem{
			Version:  matches[1],
			Name:     matches[2],
			FileName: filePath,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	fmt.Println("Found migrations:")
	for _, migration := range migrations {
		fmt.Printf("  %s (%s) in `%s`\n", migration.Version, migration.Name, migration.FileName)
	}

	dbResult, err := appDb.Query("SELECT version FROM _migrations_history ORDER BY id DESC LIMIT 1")

	var lastVersion string = ""

	if err == nil && dbResult.Next() {
		dbResult.Scan(&lastVersion)
		fmt.Printf("Last migration version: %s\n", lastVersion)
	} else {
		fmt.Printf("No migration history found, the migration will setup database.\n")
	}

	fromIdx := 0

	for i, migration := range migrations {
		if migration.Version == lastVersion {
			fromIdx = i + 1
			break
		}
	}

	if lastVersion != "" && fromIdx == 0 {
		fmt.Printf("Migration history is not found, please check the migration files.\n")
		return
	}

	if fromIdx == len(migrations) {
		fmt.Printf("Database is up to date, no migration needed.\n")
		return
	}

	fmt.Printf("Following migrations will be applied:\n")
	for i := fromIdx; i < len(migrations); i++ {
		fmt.Printf("  %s (%s)\n", migrations[i].Version, migrations[i].Name)
	}
	fmt.Printf("Do you want to continue? (y/n): ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" {
		fmt.Printf("Migration cancelled\n")
		return
	}

	for i := fromIdx; i < len(migrations); i++ {
		migration := migrations[i]

		fmt.Printf("Applying migration %s (%s)...\n", migration.Version, migration.Name)

		file, err := os.ReadFile(migration.FileName)
		if err != nil {
			fmt.Printf("Failed to read migration file: %v\n", err)
			return
		}

		_, err = appDb.Exec(string(file))
		if err != nil {
			fmt.Printf("Failed to execute migration: %v\n", err)
			return
		}

		_, err = appDb.Exec("INSERT INTO _migrations_history (version, name, time) VALUES ($1, $2, $3)", migration.Version, migration.Name, time.Now())

		if err != nil {
			fmt.Printf("Failed to update migration history: %v\n", err)
			return
		}

		fmt.Printf("Migration %s (%s) applied successfully\n", migration.Version, migration.Name)
	}

}
