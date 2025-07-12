// This package contains Sql utility functions for database operations
package sqlutil

import (
	"database/sql"
	"fmt"
	"iter"
	"reflect"
	"regexp"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
)

func camelToSnake(s string) string {
	pattern := regexp.MustCompile("(\\p{Lu}+\\P{Lu}*)")
	s = pattern.ReplaceAllString(s, "${1}_")
	s, _ = strings.CutSuffix(strings.ToLower(s), "_")
	return s
}

func getFieldColumnName(f reflect.StructField) string {
	column := f.Tag.Get("column")
	if column == "" {
		// transform camel case to snake case
		column = camelToSnake(f.Name)
	}
	return column
}

type fieldInfo struct {
	idx         int
	isPk        bool
	isGenerated bool
}

type columnToFieldInfo map[string]fieldInfo

var typeColumnToFieldIdxMapCache = make(map[reflect.Type]columnToFieldInfo)

func getTypeColumnToFieldInfo(t reflect.Type) columnToFieldInfo {
	// if t is pointer, get the element type
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// first check cache
	if ret, ok := typeColumnToFieldIdxMapCache[t]; ok {
		return ret
	}

	// build map
	ret := make(columnToFieldInfo)

	pkFound := false

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() != false && field.Tag.Get("ignore") != "true" {
			f := fieldInfo{
				idx: i,
			}
			if field.Tag.Get("pk") == "true" {
				f.isPk = true
				pkFound = true
			}
			if field.Tag.Get("generated") == "true" {
				f.isGenerated = true
			}
			ret[getFieldColumnName(field)] = f
		}
	}

	if pkFound == false {
		// if no pk found, find if id field exists
		if f, ok := ret["id"]; ok {
			ret["id"] = fieldInfo{
				idx:         f.idx,
				isPk:        true,
				isGenerated: f.isGenerated,
			}
		}
	}
	// save to cache
	typeColumnToFieldIdxMapCache[t] = ret
	return ret
}

var typePrimaryKeyCache = make(map[reflect.Type][]string)

func getTypePrimaryKey(t reflect.Type) []string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if ret, ok := typePrimaryKeyCache[t]; ok {
		return ret
	}

	cToFMap := getTypeColumnToFieldInfo(t)

	pkColumn := make([]string, 0)

	for k, v := range cToFMap {
		if v.isPk {
			pkColumn = append(pkColumn, k)
		}
	}
	typePrimaryKeyCache[t] = pkColumn
	return pkColumn
}

func rowsToEntity[T any](rows *sql.Rows) (*T, error) {
	var val T
	reflectType := reflect.TypeOf(val)
	reflectVal := reflect.New(reflectType).Elem()
	addressList := make([]interface{}, 0)

	cToFMap := getTypeColumnToFieldInfo(reflectType)

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for _, columnName := range columnNames {
		fieldIdx, ok := cToFMap[columnName]
		if !ok { // ignore column
			var ignore interface{}
			addressList = append(addressList, &ignore)
			continue
		}
		fieldVal := reflectVal.Field(fieldIdx.idx)
		newObj := reflect.New(fieldVal.Type().Elem())
		fieldVal.Set(newObj)
		addressList = append(addressList, newObj.Interface())
	}

	err = rows.Scan(addressList...)
	if err != nil {
		return nil, err
	}
	return reflectVal.Addr().Interface().(*T), nil
}

func createIterator[T any](rows *sql.Rows) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		defer rows.Close()
		for rows.Next() {
			pval, err := rowsToEntity[T](rows)
			if err != nil {
				return
			}
			if !yield(pval) {
				return
			}
		}
	}
}

func getBatchInsertQueryAndArgs[T any](tableName string, data []*T) (string, []interface{}, error) {
	if len(data) == 0 {
		return "", nil, fmt.Errorf("empty data")
	}

	reflectType := reflect.TypeOf(data[0]).Elem()
	cToFMap := getTypeColumnToFieldInfo(reflectType)

	columns := make([]string, 0)
	for k, v := range cToFMap {
		if k == "id" || k == "update_time" || v.isGenerated {
			continue
		}
		columns = append(columns, k)
	}
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no column to insert")
	}

	var values []interface{}
	for _, item := range data {
		elem := reflect.ValueOf(item).Elem()
		for _, col := range columns {
			fieldInfo := cToFMap[col]
			fieldVal := elem.Field(fieldInfo.idx)
			if fieldVal.IsNil() {
				values = append(values, nil)
			} else {
				values = append(values, fieldVal.Elem().Interface())
			}
		}
	}

	numPerData := len(columns)
	var placeholders []string
	for i := 0; i < len(data); i++ {
		ph := make([]string, numPerData)
		start := i*numPerData + 1
		for j := 0; j < numPerData; j++ {
			ph[j] = fmt.Sprintf("$%d", start+j)
		}
		placeholders = append(placeholders, "("+strings.Join(ph, ", ")+")")
	}

	columnsStr := strings.Join(columns, ", ")
	valuesStr := strings.Join(placeholders, ", ")
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		columnsStr,
		valuesStr,
	)

	return insertSQL, values, nil
}
func getInsertQueryAndArgs[T any](tableName string, data *T, returning bool) (string, []interface{}, error) {
	reflectType := reflect.TypeOf(*data)
	reflectVal := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	returningColumns := make([]string, 0)
	values := make([]interface{}, 0)

	cToFMap := getTypeColumnToFieldInfo(reflectType)

	for k, v := range cToFMap {
		if v.isGenerated {
			returningColumns = append(returningColumns, k)
		}
		// if generated or nil, ignore
		if v.isGenerated || reflectVal.Field(v.idx).IsNil() {
			continue
		}
		columns = append(columns, k)
		values = append(values, reflectVal.Field(v.idx).Elem().Interface())
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no column to insert")
	}

	columnsStr := strings.Join(columns, ", ")
	returningColumnsStr := strings.Join(returningColumns, ", ")
	// value str $1, $2
	valuesArr := make([]string, 0)
	for i := 1; i <= len(values); i++ {
		valuesArr = append(valuesArr, fmt.Sprintf("$%d", i))
	}
	valuesStr := strings.Join(valuesArr, ", ")

	insertSentenceTemplate := `INSERT INTO %s (%s) VALUES (%s)`
	var insertSentence string
	if returning && returningColumnsStr != "" {
		insertSentenceTemplate += ` RETURNING %s`
		insertSentence = fmt.Sprintf(insertSentenceTemplate, tableName, columnsStr, valuesStr, returningColumnsStr)
	} else {
		insertSentence = fmt.Sprintf(insertSentenceTemplate, tableName, columnsStr, valuesStr)
	}

	return insertSentence, values, nil
}

func getUpdateQueryAndArgs[T any](tableName string, data *T) (string, []interface{}, error) {
	reflectType := reflect.TypeOf(*data)
	reflectVal := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	values := make([]interface{}, 0)

	whereColumns := make([]string, 0)
	whereValues := make([]interface{}, 0)

	cToFMap := getTypeColumnToFieldInfo(reflectType)
	pkColumns := getTypePrimaryKey(reflectType)

	for k, v := range cToFMap {
		// if pk or generated or nil, ignore
		if v.isPk || v.isGenerated || reflectVal.Field(v.idx).IsNil() {
			continue
		}
		columns = append(columns, k)
		values = append(values, reflectVal.Field(v.idx).Elem().Interface())
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no column to update")
	}

	queryPlacement := 1
	valuesArr := make([]string, 0)
	for _, col := range columns {
		valuesArr = append(valuesArr, fmt.Sprintf("%s = $%d", col, queryPlacement))
		queryPlacement++
	}
	valuesStr := strings.Join(valuesArr, ", ")

	whereArr := make([]string, 0)
	for _, pkColumn := range pkColumns {
		whereArr = append(whereArr, fmt.Sprintf("%s = $%d", pkColumn, queryPlacement))
		queryPlacement++
	}
	whereStr := strings.Join(whereArr, " AND ")

	for _, pkColumn := range pkColumns {
		fieldInfo, ok := cToFMap[pkColumn]
		if !ok {
			return "", nil, fmt.Errorf("primary key column %s not found in struct", pkColumn)
		}
		whereColumns = append(whereColumns, pkColumn)
		if reflectVal.Field(fieldInfo.idx).IsNil() {
			return "", nil, fmt.Errorf("primary key column %s is nil", pkColumn)
		}
		whereValues = append(whereValues, reflectVal.Field(fieldInfo.idx).Elem().Interface())
	}

	values = append(values, whereValues...)

	updateSentenceTemplate := `UPDATE %s SET %s WHERE %s`
	updateSentence := fmt.Sprintf(updateSentenceTemplate, tableName, valuesStr, whereStr)

	return updateSentence, values, nil
}

func getUpsertQueryAndArgs[T any](tableName string, data *T) (string, []interface{}, error) {
	reflectType := reflect.TypeOf(*data)
	reflectVal := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	pks := make([]string, 0)
	values := make([]interface{}, 0)

	returningColumns := make([]string, 0)

	cToFMap := getTypeColumnToFieldInfo(reflectType)

	for k, v := range cToFMap {
		if v.isGenerated {
			returningColumns = append(returningColumns, k)
		}
		// if generated or nil, ignore
		if v.isGenerated || reflectVal.Field(v.idx).IsNil() {
			continue
		}
		if v.isPk {
			pks = append(pks, k)
		}
		columns = append(columns, k)
		values = append(values, reflectVal.Field(v.idx).Elem().Interface())
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no column to insert")
	}

	pkStr := strings.Join(pks, ", ")

	columnsStr := strings.Join(columns, ", ")
	// value str $1, $2
	valuesArr := make([]string, 0)
	for i := 1; i <= len(values); i++ {
		valuesArr = append(valuesArr, fmt.Sprintf("$%d", i))
	}
	valuesStr := strings.Join(valuesArr, ", ")

	updateStr := ""

	returningColumnsStr := strings.Join(returningColumns, ", ")

	for i, column := range columns {
		if i != 0 {
			updateStr += ", "
		}
		updateStr += fmt.Sprintf("%s = $%d", column, i+1)
	}

	insertOrUpdateSentenceTemplate := `INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s`

	var insertOrUpdateSentence string

	if returningColumnsStr != "" {
		insertOrUpdateSentenceTemplate += ` RETURNING %s`
		insertOrUpdateSentence = fmt.Sprintf(insertOrUpdateSentenceTemplate, tableName, columnsStr, valuesStr, pkStr, updateStr, returningColumnsStr)
	} else {
		insertOrUpdateSentence = fmt.Sprintf(insertOrUpdateSentenceTemplate, tableName, columnsStr, valuesStr, pkStr, updateStr)
	}
	return insertOrUpdateSentence, values, nil
}

func getDeleteQueryAndArgs[T any](tableName string, data *T) (string, []interface{}, error) {
	reflectType := reflect.TypeOf(*data)
	reflectVal := reflect.ValueOf(data).Elem()

	whereColumns := make([]string, 0)
	whereValues := make([]interface{}, 0)

	cToFMap := getTypeColumnToFieldInfo(reflectType)
	pkColumns := getTypePrimaryKey(reflectType)

	queryPlacement := 1
	whereArr := make([]string, 0)
	for _, pkColumn := range pkColumns {
		whereArr = append(whereArr, fmt.Sprintf("%s = $%d", pkColumn, queryPlacement))
		queryPlacement++
	}
	whereStr := strings.Join(whereArr, " AND ")

	for _, pkColumn := range pkColumns {
		fieldInfo, ok := cToFMap[pkColumn]
		if !ok {
			return "", nil, fmt.Errorf("primary key column %s not found in struct", pkColumn)
		}
		whereColumns = append(whereColumns, pkColumn)
		if reflectVal.Field(fieldInfo.idx).IsNil() {
			return "", nil, fmt.Errorf("primary key column %s is nil", pkColumn)
		}
		whereValues = append(whereValues, reflectVal.Field(fieldInfo.idx).Elem().Interface())
	}

	deleteSentenceTemplate := `DELETE FROM %s WHERE %s`
	deleteSentence := fmt.Sprintf(deleteSentenceTemplate, tableName, whereStr)

	return deleteSentence, whereValues, nil
}

func Query[T any](ctx storage.AppDatabaseContext, query string, args ...interface{}) (iter.Seq[*T], error) {
	rows, err := ctx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return createIterator[T](rows), nil
}

func QueryFirst[T any](ctx storage.AppDatabaseContext, query string, args ...interface{}) (*T, error) {
	rows, err := ctx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		return rowsToEntity[T](rows)
	}
	return nil, nil
}

func getSelectQuery[T any](from string, afterFrom string) string {
	// data is only use for get the type of T
	data := new(T)
	reflectType := reflect.TypeOf(*data)

	cToFMap := getTypeColumnToFieldInfo(reflectType)

	fields := make([]string, 0)

	for k := range cToFMap {
		fields = append(fields, k)
	}

	query := fmt.Sprintf(`SELECT %s FROM %s %s`, strings.Join(fields, ", "), from, afterFrom)
	return query
}

func QueryCommon[T any](ctx storage.AppDatabaseContext, from string, afterFrom string, args ...interface{}) (iter.Seq[*T], error) {
	query := getSelectQuery[T](from, afterFrom)
	return Query[T](ctx, query, args...)
}

func QueryCommonFirst[T any](ctx storage.AppDatabaseContext, from string, afterFrom string, args ...interface{}) (*T, error) {
	query := getSelectQuery[T](from, afterFrom+" LIMIT 1")
	return QueryFirst[T](ctx, query, args...)
}

func scanGeneratedColumns(data interface{}, rows *sql.Rows) {
	reflectType := reflect.TypeOf(data).Elem()
	reflectVal := reflect.ValueOf(data).Elem()

	cToFMap := getTypeColumnToFieldInfo(reflectType)

	columnNames, err := rows.Columns()
	if err != nil {
		return
	}

	for _, columnName := range columnNames {
		fieldInfo, ok := cToFMap[columnName]
		if !ok {
			continue
		}
		if fieldInfo.isGenerated {
			fieldVal := reflectVal.Field(fieldInfo.idx)
			newObj := reflect.New(fieldVal.Type().Elem())
			rows.Scan(newObj.Interface())
			fieldVal.Set(newObj)
		}
	}
}

func Insert[T any](ctx storage.AppDatabaseContext, into string, data *T) error {
	insertSentence, values, err := getInsertQueryAndArgs[T](into, data, true)
	if err != nil {
		return err
	}
	rows, err := ctx.Query(insertSentence, values...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		scanGeneratedColumns(data, rows)
	}

	return err
}

func BatchInsert[T any](ctx storage.AppDatabaseContext, into string, data []*T) error {
	const BatchInsertSizePerTime = 1000

	if len(data) == 0 {
		return fmt.Errorf("no data to insert")
	}

	for i := 0; i < len(data); i += BatchInsertSizePerTime {
		d := data[i:min(i+BatchInsertSizePerTime, len(data))]

		insertSentence, values, err := getBatchInsertQueryAndArgs[T](into, d)
		if err != nil {
			return err
		}
		_, err = ctx.Exec(insertSentence, values...)
		if err != nil {
			return err
		}
	}
	return nil
}

func Update[T any](ctx storage.AppDatabaseContext, tableName string, data *T) error {
	updateSentence, values, err := getUpdateQueryAndArgs[T](tableName, data)
	if err != nil {
		return err
	}
	_, err = ctx.Exec(updateSentence, values...)
	return err
}

func Upsert[T any](ctx storage.AppDatabaseContext, into string, data *T) error {
	insertSentence, values, err := getUpsertQueryAndArgs[T](into, data)
	if err != nil {
		return err
	}
	rows, err := ctx.Query(insertSentence, values...)
	if err != nil {
		return err
	}
	if rows.Next() {
		scanGeneratedColumns(data, rows)
	}
	return err
}

func Delete[T any](ctx storage.AppDatabaseContext, tableName string, data *T) error {
	deleteSentence, values, err := getDeleteQueryAndArgs[T](tableName, data)
	if err != nil {
		return err
	}
	_, err = ctx.Exec(deleteSentence, values...)
	return err
}

// MergeStruct merge old struct to dst struct
// if field in dst is nil, set it to the value of old
func MergeStruct[T any](old *T, dst *T) {
	srcVal := reflect.ValueOf(old).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)
		if dstField.IsNil() {
			dstField.Set(srcField)
		}
	}
}

// ToData convert data to pointer
func ToData[T any](data T) *T {
	return &data
}

// ToNullable convert data to nullable pointer
func ToNullable[T any](data T) **T {
	d := &data
	return &d
}

func IsUnset[T any](data *T) bool {
	return data == nil
}

func IsNull[T any](data **T) bool {
	return data == nil || *data == nil
}
