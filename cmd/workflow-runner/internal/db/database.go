package db

import (
	"database/sql"
	"os"
	"path"
	"strings"
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/internal/workflow"
	"github.com/HUSTSecLab/criticality_score/cmd/workflow-runner/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
)

var db *sql.DB

func OpenAndInitDB() {
	// create dir if not exists
	os.MkdirAll(config.GetWorkflowHistoryDir(), 0755)

	// get sqlite db path
	dbPath := path.Join(config.GetWorkflowHistoryDir(), "history.db")
	// open db
	var err error
	db, err = sql.Open("sqlite3", dbPath)

	if err != nil {
		panic(err)
	}

	// create table if not exists
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS round (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		startTime TIMESTAMP,
		endTime TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS workflow (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		roundId INTEGER,
		name TEXT,
		title TEXT,
		description TEXT,
		args TEXT,
		status TEXT,
		type TEXT,
		dependency TEXT,
		startTime TIMESTAMP,
		endTime TIMESTAMP,
		-- foreign key constraint
		FOREIGN KEY (roundId) REFERENCES round(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS cfg (
		key TEXT PRIMARY KEY,
		value TEXT
	);`)
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

func CreateRound(tasks []*workflow.WorkflowNode) (int, error) {
	// create round
	tx, err := db.Begin()

	if err != nil {
		return 0, err
	}

	// insert round
	result, err := tx.Exec("INSERT INTO round (startTime) VALUES (datetime('now'))")
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// get last insert id
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// write all tasks to db
	for _, task := range tasks {

		deps := strings.Join(lo.Map(task.Dependencies, func(dep *workflow.WorkflowNode, _ int) string {
			return dep.Name
		}), ",")

		_, err := tx.Exec(`
		INSERT INTO workflow (roundId, name, title, description, args, status, type, dependency)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, id, task.Name, task.Title, task.Description, task.DefaultArgs, rpc.TaskStatusPending, task.Type, deps)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetRound(id int) (*rpc.RoundDTO, error) {
	var round rpc.RoundDTO
	err := db.QueryRow(`
	SELECT id, startTime, endTime
	FROM round
	WHERE id = ?
	`, id).Scan(&round.ID, &round.StartTime, &round.EndTime)
	if err != nil {
		return nil, err
	}
	// get all tasks
	rows, err := db.Query(`
	SELECT name, title, description, args, status, type, dependency, startTime, endTime
	FROM workflow
	WHERE roundId = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	round.Tasks = make([]rpc.TaskDTO, 0)

	// scan all tasks
	for rows.Next() {
		var task rpc.TaskDTO
		var deps string
		err := rows.Scan(&task.Name, &task.Title, &task.Description, &task.Args, &task.Status, &task.Type, &deps, &task.StartTime, &task.EndTime)
		if err != nil {
			return nil, err
		}
		// split deps
		task.Dependencies = lo.Map(strings.Split(deps, ","), func(dep string, _ int) string {
			return dep
		})
		// append task to round
		round.Tasks = append(round.Tasks, task)
	}
	// check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &round, nil
}

func GetMaxRoundID() (int, error) {
	var maxID int
	err := db.QueryRow(`
	SELECT MAX(id)
	FROM round
	`).Scan(&maxID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // no rounds found
		}
		return 0, err
	}
	return maxID, nil
}

func GetTask(roundID int, taskName string) (*rpc.TaskDTO, error) {
	var task rpc.TaskDTO
	var deps string
	err := db.QueryRow(`
	SELECT name, title, description, args, status, type, dependency, startTime, endTime
	FROM workflow
	WHERE roundId = ? AND name = ?
	`, roundID, taskName).Scan(&task.Name, &task.Title, &task.Description, &task.Args, &task.Status, &task.Type, &deps, &task.StartTime, &task.EndTime)
	if err != nil {
		return nil, err
	}
	// split deps
	task.Dependencies = lo.Map(strings.Split(deps, ","), func(dep string, _ int) string {
		return dep
	})
	return &task, nil
}

func UpdateTask(roundID int, task *rpc.TaskDTO) error {
	// update task
	_, err := db.Exec(`
	UPDATE workflow
	SET status = ?, startTime = ?, endTime = ?, args = ?
	WHERE name = ? AND roundId = ?
	`, task.Status, task.StartTime, task.EndTime, task.Args, task.Name, roundID)
	if err != nil {
		return err
	}
	return nil
}

func GetLastTriggerTime(taskName string) (start, end time.Time, err error) {
	// get last trigger time
	err = db.QueryRow(`
	SELECT startTime, endTime
	FROM workflow
	WHERE name = ?
	ORDER BY startTime DESC
	LIMIT 1
	`, taskName).Scan(&start, &end)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, time.Time{}, nil
		}
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func GetConfig(key string) (string, error) {
	// get config value
	var value string
	err := db.QueryRow(`
	SELECT value
	FROM cfg
	WHERE key = ?
	`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // return empty string if not found
		}
		return "", err
	}
	return value, nil
}

func SetConfig(key, value string) error {
	// insert or update config value
	_, err := db.Exec(`
	INSERT INTO cfg (key, value)
	VALUES (?, ?)
	ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	if err != nil {
		return err
	}
	return nil
}
