package repository

import (
	"iter"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/sqlutil"
)

type WorkflowHistoryRepository interface {
	/** QUERY **/
	QueryByJobID(jobID string) (iter.Seq[*WorkflowHistory], error)

	GetLatestJobID() (string, error)
	GetLatestByTaskName(taskName string) (*WorkflowHistory, error)

	/** INSERT/UPDATE **/
	Insert(data *WorkflowHistory) error
}

const (
	WorkflowHistoryActionStart = "start"
	WorkflowHistoryActionEnd   = "end"
	WorkflowHistoryActionError = "error"
)

type WorkflowHistory struct {
	ID       *int64 `pk:"true" generated:"true"`
	JobID    *string
	TaskName *string
	Action   *string
	Payload  *string
}

const WorkflowHistoryTableName = "workflows"

type workflowHistoryRepository struct {
	ctx storage.AppDatabaseContext
}

var _ WorkflowHistoryRepository = (*workflowHistoryRepository)(nil)

func NewWorkflowHistoryRepository(ctx storage.AppDatabaseContext) WorkflowHistoryRepository {
	return &workflowHistoryRepository{ctx: ctx}
}

func (w *workflowHistoryRepository) GetLatestByTaskName(taskName string) (*WorkflowHistory, error) {
	return sqlutil.QueryCommonFirst[WorkflowHistory](w.ctx, WorkflowHistoryTableName,
		"WHERE task_name = $1 ORDER BY id DESC LIMIT 1", taskName)
}

func (w *workflowHistoryRepository) GetLatestJobID() (string, error) {
	ctx := w.ctx
	row := ctx.QueryRow(`SELECT job_id FROM workflow_history ORDER BY id DESC LIMIT 1`)

	var jobID string
	err := row.Scan(&jobID)
	if err != nil {
		return "", err
	}
	return jobID, nil
}

func (w *workflowHistoryRepository) Insert(data *WorkflowHistory) error {
	if data.JobID == nil || *data.JobID == "" || data.TaskName == nil || *data.TaskName == "" {
		return ErrInvalidInput
	}

	return sqlutil.Insert(w.ctx, WorkflowHistoryTableName, data)
}

func (w *workflowHistoryRepository) QueryByJobID(jobID string) (iter.Seq[*WorkflowHistory], error) {
	return sqlutil.QueryCommon[WorkflowHistory](w.ctx, WorkflowHistoryTableName, "WHERE job_id = $1", jobID)
}
