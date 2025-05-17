package task

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/git"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/util"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/repository"
	"github.com/HUSTSecLab/criticality_score/pkg/storage/sqlutil"
	"github.com/lib/pq"
)

type RunningTask struct {
	Start    time.Time
	Progress TermBuffer
}

var runningTasks = make(map[string]*RunningTask, 0)
var muRunningTasks sync.Mutex

func GetRunningTasks() []rpc.RunningTaskDTO {
	muRunningTasks.Lock()
	defer muRunningTasks.Unlock()

	tasks := make([]rpc.RunningTaskDTO, 0, len(runningTasks))
	for k, v := range runningTasks {
		tasks = append(tasks, rpc.RunningTaskDTO{
			Link:     k,
			Start:    v.Start,
			Progress: v.Progress.String(),
		})
	}
	return tasks
}

func Collect(gitLink string, disableCollect bool) {
	timeBegin := time.Now()

	var currentTask = RunningTask{
		Start:    timeBegin,
		Progress: TermBuffer{},
	}

	muRunningTasks.Lock()
	runningTasks[gitLink] = &currentTask
	muRunningTasks.Unlock()

	defer func() {
		muRunningTasks.Lock()
		delete(runningTasks, gitLink)
		muRunningTasks.Unlock()
	}()

	// TODO: consider remove this recover function
	defer func() {
		if p := recover(); p != nil {
			logger.WithFields(map[string]any{
				"gitLink": gitLink,
			}).Errorf("PANIC occurs: %v", p)
		}
	}()

	gmr := repository.NewGitMetricsRepository(storage.GetDefaultAppDatabaseContext())

	gf, err := gmr.GetGitFileByLink(gitLink)
	// begin get file path
	var filePathRel string
	var filePathAbs string

	if gf != nil && gf.FilePath != nil && err == nil {
		filePathRel = *gf.FilePath
	}

	pathNotExists := func(p string) bool {
		_, err := os.Stat(p)
		return errors.Is(err, os.ErrNotExist)
	}

	if filePathRel == "" || pathNotExists(filepath.Join(config.GetGitStoragePath(), filePathRel)) {
		if filePathRel != "" {
			logger.WithFields(map[string]any{
				"oldpath": filePathRel,
				"gitlink": gitLink,
			}).Warnf("file path in database is not exsits in filesystem, regenerate again")
		}
		filePathRel, err = util.GetGitRepositoryPathFromURL("", gitLink)
		if err != nil {
			logger.Errorf("could not parse url: %v", err)
			return
		}
	}

	filePathAbs, err = filepath.Abs(filepath.Join(config.GetGitStoragePath(), filePathRel))
	if err != nil {
		logger.Errorf("filepath generate fail")
		return
	}

	recordClone := func(success bool, e error) {
		if !success {
			logger.WithFields(map[string]any{
				"gitlink": gitLink,
				"error":   e,
			}).Errorf("Clone git metrics failed: %v", gitLink)
		}
		var msg **string
		if e != nil {
			msg = sqlutil.ToNullable(e.Error() + "\nOutput:\n" + currentTask.Progress.String())
		} else {
			msg = sqlutil.ToData[*string](nil)
		}

		err := gmr.InsertOrUpdateGitFile(&repository.GitFile{
			GitLink:    sqlutil.ToData(gitLink),
			FilePath:   sqlutil.ToData(filePathRel),
			Message:    msg,
			UpdateTime: sqlutil.ToNullable(time.Now()),
			TakeTimeMs: sqlutil.ToNullable(time.Since(timeBegin).Milliseconds()),
		}, success)
		if err != nil {
			logger.WithFields(map[string]any{
				"gitlink": gitLink,
				"error":   err,
			}).Errorf("Inserting row failed: %v", gitLink)
		}
	}

	recordParseSuccess := func(repo *git.Repo) {
		// logger.WithFields(map[string]any{
		// 	"gitlink": gitLink,
		// }).Infof("git metrics collected successfully: %v", gitLink)

		err := gmr.InsertOrUpdate(&repository.GitMetric{
			GitLink:          sqlutil.ToData(gitLink),
			CreatedSince:     sqlutil.ToNullable(repo.CreatedSince),
			UpdatedSince:     sqlutil.ToNullable(repo.UpdatedSince),
			ContributorCount: sqlutil.ToNullable(repo.ContributorCount),
			CommitFrequency:  sqlutil.ToNullable(repo.CommitFrequency),
			OrgCount:         sqlutil.ToNullable(repo.OrgCount),
			//* License:          sqlutil.ToNullable(pq.StringArray(repo.Licenses)),
			Language: sqlutil.ToNullable(pq.StringArray(repo.Languages)),
		})

		if err != nil {
			logger.Errorf("Inserting %s Failed", gitLink)
		}

	}

	u, err := url.ParseURL(gitLink)
	if err != nil {
		logger.Errorf("url.ParseURL fail: %s: %v", gitLink, err)
		return
	}
	r, err := collector.Collect(&u, filePathAbs, &currentTask.Progress)
	if err != nil {
		recordClone(false, err)
		return
	}
	recordClone(true, nil)

	if !disableCollect {
		repo, err := git.ParseRepo(r)
		if err != nil {
			logger.WithFields(map[string]any{
				"gitlink": gitLink,
			}).Errorf("Parse repo error: %v", err)
			return
		}
		recordParseSuccess(repo)
	}
}
