package tool

import (
	"database/sql"
	"os"
	"path"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenAndInitDB() {
	// create dir if not exists
	os.MkdirAll(config.GetWebToolHistoryDir(), 0755)

	// get sqlite db path
	dbPath := path.Join(config.GetWebToolHistoryDir(), "history.db")
	// open db
	var err error
	db, err = sql.Open("sqlite3", dbPath)

	if err != nil {
		panic(err)
	}

	// create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id STRING PRIMARY KEY,
			toolId STRING,
			toolName STRING,
			launchUserName STRING,
			startTime TIMESTAMP,
			endTime TIMESTAMP,
			ret INTEGER,
			err STRING
		);
		CREATE INDEX IF NOT EXISTS idx_toolId ON history (toolId);
		CREATE INDEX IF NOT EXISTS idx_launchUserName ON history (launchUserName);
		CREATE INDEX IF NOT EXISTS idx_startTime ON history (startTime);
		CREATE INDEX IF NOT EXISTS idx_endTime ON history (endTime);
		CREATE INDEX IF NOT EXISTS idx_ret ON history (ret);
		CREATE INDEX IF NOT EXISTS idx_err ON history (err);
	`)
	if err != nil {
		panic(err)
	}
}

func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}
}

func SaveToolInstanceHistory(instance *ToolInstanceHistory) error {
	// upsert into db
	_, err := db.Exec(`
		INSERT INTO history (id, toolId, toolName, launchUserName, startTime, endTime, ret, err)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			toolId = excluded.toolId,
			toolName = excluded.toolName,
			launchUserName = excluded.launchUserName,
			startTime = excluded.startTime,
			endTime = excluded.endTime,
			ret = excluded.ret,
			err = excluded.err
	`, instance.ID, instance.ToolID, instance.ToolName, instance.LaunchUserName, instance.StartTime, instance.EndTime, instance.Ret, instance.Err)

	return err
}

func QueryToolInstanceHistory(id string) (*ToolInstanceHistory, error) {
	// select from db
	row := db.QueryRow(`
		SELECT id, toolId, toolName, launchUserName, startTime, endTime, ret, err
		FROM history
		WHERE id = ?
	`, id)

	var instance ToolInstanceHistory
	err := row.Scan(&instance.ID, &instance.ToolID, &instance.ToolName, &instance.LaunchUserName, &instance.StartTime, &instance.EndTime, &instance.Ret, &instance.Err)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func QueryToolInstancesHistoryOrderByStartTime(skip, take int) ([]*ToolInstanceHistory, error) {
	// select from db
	rows, err := db.Query(`
		SELECT id, toolId, toolName, launchUserName, startTime, endTime, ret, err
		FROM history
		ORDER BY startTime DESC
		LIMIT ? OFFSET ?
	`, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*ToolInstanceHistory
	for rows.Next() {
		var instance ToolInstanceHistory
		err := rows.Scan(&instance.ID, &instance.ToolID, &instance.ToolName, &instance.LaunchUserName, &instance.StartTime, &instance.EndTime, &instance.Ret, &instance.Err)
		if err != nil {
			return nil, err
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

func CountToolInstanceHistories() (int, error) {
	// count from db
	row := db.QueryRow(`
		SELECT COUNT(*)
		FROM history
	`)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
