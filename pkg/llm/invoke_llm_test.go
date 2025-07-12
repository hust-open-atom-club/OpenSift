package llm_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HUSTSecLab/OpenSift/pkg/llm"
)

func TestUpdateBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	resultMap := map[string]map[string]string{
		"repo1": {
			"package1": "https://gitlink1.com",
			"package2": "https://gitlink2.com",
		},
		"repo2": {
			"package3": "https://gitlink3.com",
		},
	}

	batchSize := 2

	mock.ExpectExec(`UPDATE repo1 SET git_link = CASE`).
		WithArgs("repo1", "package1", "https://gitlink1.com", "package2", "https://gitlink2.com").
		WillReturnResult(sqlmock.NewResult(1, 2))

	mock.ExpectExec(`UPDATE repo2 SET git_link = CASE`).
		WithArgs("repo2", "package3", "https://gitlink3.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = llm.UpdateBatch(db, batchSize, resultMap)
	if err != nil {
		t.Errorf("UpdateBatch failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestUpdateIdxBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	gitIndustry := map[string]string{
		"https://gitlink1.com": "industry1",
		"https://gitlink2.com": "industry2",
		"https://gitlink3.com": "industry3",
	}

	batchSize := 2

	mock.ExpectExec("UPDATE git_repositories SET industry = CASE").
		WithArgs("https://gitlink1.com", "industry1", "https://gitlink2.com", "industry2").
		WillReturnResult(sqlmock.NewResult(1, 2))

	err = llm.UpdateIdxBatch(db, batchSize, gitIndustry)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}
