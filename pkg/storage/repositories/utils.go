package repositories

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

type DatabaseRepositoryIterator[T any] struct {
	rows *sql.Rows
}

func newDatabaseRepositoryIterator[T any](rows *sql.Rows) *DatabaseRepositoryIterator[T] {
	return &DatabaseRepositoryIterator[T]{rows: rows}
}

func (i *DatabaseRepositoryIterator[T]) Close() error {
	return i.rows.Close()
}

func (i *DatabaseRepositoryIterator[T]) Next() (*T, error) {

	rows := i.rows

	if !rows.Next() {
		return nil, nil
	}

	val := newDataWithMakeEmpty[T]()

	reflectType := reflect.TypeOf(val).Elem()

	data := make([]interface{}, 0)

	for i := 0; i < reflectType.NumField(); i++ {
		if reflectType.Field(i).Tag.Get("column") != "" {
			data = append(data, reflect.ValueOf(val).Elem().Field(i).Addr().Interface())
		}
	}

	err := rows.Scan(data...)

	if err != nil {
		return nil, err
	}

	return val, nil
}

func newDataWithMakeEmpty[T any]() *T {
	data := new(T)
	// make all fields a address to avoid nil pointer using reflection
	reflectType := reflect.TypeOf(*data)
	reflectValue := reflect.ValueOf(data).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectValue.Field(i)
		if field.CanSet() {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}
	return data
}

func getDataFromTable[T any](appDb *storage.AppDatabase, tableName string, whereAndOrderBySentence string, args ...interface{}) (*DatabaseRepositoryIterator[T], error) {
	// data is only use for get the type of T
	data := new(T)

	reflectType := reflect.TypeOf(*data)
	selectFields := make([]string, 0)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if field.Tag.Get("column") != "" {
			selectFields = append(selectFields, field.Tag.Get("column"))
		}
	}
	selectStr := strings.Join(selectFields, ", ")

	qeury := fmt.Sprintf(`SELECT %s FROM %s %s`, selectStr, tableName, whereAndOrderBySentence)

	rows, err := appDb.Query(qeury, args...)
	if err != nil {
		return nil, err
	}

	return newDatabaseRepositoryIterator[T](rows), nil
}

func getInsertQueryAndArgs[T any](tableName string, data *T) (string, []interface{}) {
	reflectType := reflect.TypeOf(*data)
	dataReflectVal := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	values := make([]interface{}, 0)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		if field.Tag.Get("column") == "" ||
			field.Tag.Get("readonly") == "true" {
			continue
		}

		if dataReflectVal.Field(i).IsNil() {
			continue
		}

		columns = append(columns, field.Tag.Get("column"))
		values = append(values, dataReflectVal.Field(i).Interface())
	}

	columnsStr := strings.Join(columns, ", ")
	// value str $1, $2
	valuesArr := make([]string, 0)
	for i := 1; i <= len(values); i++ {
		valuesArr = append(valuesArr, fmt.Sprintf("$%d", i))
	}
	valuesStr := strings.Join(valuesArr, ", ")

	insertSentenceTemplate := `INSERT INTO %s (%s) VALUES (%s)`

	insertSentence := fmt.Sprintf(insertSentenceTemplate, tableName, columnsStr, valuesStr)

	return insertSentence, values

}

func insertDataIntoTable[T any](appDb *storage.AppDatabase, tableName string, data *T) error {
	insertSentence, values := getInsertQueryAndArgs[T](tableName, data)
	_, err := appDb.Exec(insertSentence, values...)
	return err
}

func batchInsertDataIntoTable[T any](appDb *storage.AppDatabase, tableName string, data []*T) error {

	batchCtx := appDb.NewBatchExecContext(&storage.BatchExecContextConfig{
		AutoCommit:     true,
		AutoCommitSize: 1000,
	})

	for _, d := range data {
		insertSentence, args := getInsertQueryAndArgs[T](tableName, d)
		batchCtx.AppendExec(insertSentence, args...)
	}

	batchCtx.Commit()
	return nil
}
