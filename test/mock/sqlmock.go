// Package mock provides SQL mocking utilities for repository unit tests.
// This package uses gomock-style interface mocking for testing repository layers.
package mock

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBTXMock is a mock implementation of the DBTX interface for testing.
type DBTXMock struct {
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Exec mocks database execution.
func (m *DBTXMock) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return pgconn.NewCommandTag("DELETE 1"), nil
}

// Query mocks database query execution.
func (m *DBTXMock) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

// QueryRow mocks single row query execution.
func (m *DBTXMock) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

// RowsMock is a mock implementation of pgx.Rows for testing.
type RowsMock struct {
	err  error
	conn *pgx.Conn
	rows [][]any
	idx  int
}

// NewRowsMock creates a new RowsMock with the given row data.
func NewRowsMock(rows [][]any) *RowsMock {
	return &RowsMock{rows: rows, idx: -1, conn: nil}
}

// Close closes the rows mock.
func (r *RowsMock) Close() {}

// Err returns any error that occurred during iteration.
func (r *RowsMock) Err() error { return r.err }

// CommandTag returns the command tag for the query.
func (r *RowsMock) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("SELECT %d", len(r.rows)))
}

// Conn returns the underlying connection.
func (r *RowsMock) Conn() *pgx.Conn { return r.conn }

// Next advances to the next row.
func (r *RowsMock) Next() bool {
	r.idx++
	return r.idx < len(r.rows)
}

// Scan scans the current row into the given destinations.
func (r *RowsMock) Scan(dest ...any) error {
	rowData := r.getRowData()
	if rowData == nil {
		return nil
	}

	if !r.scanStruct(dest, rowData) {
		r.scanFields(dest, rowData)
	}
	return nil
}

// getRowData returns the current row data or nil if out of bounds.
func (r *RowsMock) getRowData() []any {
	if r.idx < 0 || r.idx >= len(r.rows) {
		return nil
	}
	return r.rows[r.idx]
}

// scanStruct attempts to scan into a single struct pointer.
func (r *RowsMock) scanStruct(dest, rowData []any) bool {
	if len(dest) != 1 || len(rowData) <= 1 {
		return false
	}

	destVal := reflect.ValueOf(dest[0])
	if !isStructPointer(destVal) {
		return false
	}

	structVal := destVal.Elem()
	minFields := min(len(rowData), structVal.NumField())

	for i := 0; i < minFields; i++ {
		fieldVal := structVal.Field(i)
		if fieldVal.CanSet() {
			_ = scanValue(fieldVal.Addr().Interface(), rowData[i])
		}
	}
	return true
}

// isStructPointer checks if the value is a pointer to a struct.
func isStructPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct
}

// scanFields scans each field individually.
func (r *RowsMock) scanFields(dest, rowData []any) {
	for i := range rowData {
		if i >= len(dest) {
			break
		}
		_ = scanValue(dest[i], rowData[i])
	}
}

// FieldDescriptions returns the field descriptions for the rows.
func (r *RowsMock) FieldDescriptions() []pgconn.FieldDescription {
	return make([]pgconn.FieldDescription, 0)
}

// Values returns the values of the current row.
func (r *RowsMock) Values() ([]any, error) {
	rowData := r.getRowData()
	if rowData == nil {
		return nil, nil
	}
	return rowData, nil
}

// RawValues returns the raw byte values of the current row.
func (r *RowsMock) RawValues() [][]byte {
	rowData := r.getRowData()
	if rowData == nil {
		return nil
	}
	raw := make([][]byte, len(rowData))
	for i, v := range rowData {
		if s, ok := v.(string); ok {
			raw[i] = []byte(s)
		}
	}
	return raw
}

// RowMock is a mock implementation of pgx.Row for testing.
type RowMock struct {
	err error
	row []any
}

// NewRowMock creates a new RowMock with the given row data.
func NewRowMock(row []any) *RowMock {
	return &RowMock{row: row}
}

// Scan scans the row into the given destinations.
func (r *RowMock) Scan(dest ...any) error {
	if len(r.row) == 0 {
		return pgx.ErrNoRows
	}

	if r.scanStructDest(dest) {
		return r.err
	}

	return r.scanFields(dest)
}

// scanStructDest handles case where dest is a single struct pointer.
func (r *RowMock) scanStructDest(dest []any) bool {
	if len(dest) != 1 {
		return false
	}

	destVal := reflect.ValueOf(dest[0])
	if !isStructPointer(destVal) {
		return false
	}

	destType := destVal.Elem().Type()

	if len(r.row) > 1 {
		r.scanStructFields(destVal.Elem(), destType)
		return true
	}

	if len(r.row) == 1 {
		if r.scanSingleStructValue(destVal.Elem(), destType) {
			return true
		}
	}

	return false
}

// scanStructFields scans individual field values into struct fields.
func (r *RowMock) scanStructFields(structVal reflect.Value, structType reflect.Type) {
	minFields := min(len(r.row), structVal.NumField())
	for i := 0; i < minFields; i++ {
		fieldVal := structVal.Field(i)
		if fieldVal.CanSet() {
			_ = scanValue(fieldVal.Addr().Interface(), r.row[i])
		}
	}
}

// scanSingleStructValue handles direct struct assignment from single row value.
func (r *RowMock) scanSingleStructValue(destVal reflect.Value, destType reflect.Type) bool {
	rowVal := reflect.ValueOf(r.row[0])
	if rowVal.Kind() != reflect.Struct {
		return false
	}
	if !rowVal.Type().AssignableTo(destType) {
		return false
	}
	destVal.Set(rowVal)
	return true
}

// scanFields scans each field individually.
func (r *RowMock) scanFields(dest []any) error {
	for i := range r.row {
		if i >= len(dest) {
			break
		}
		if err := scanValue(dest[i], r.row[i]); err != nil {
			return err
		}
	}
	return r.err
}

// scanValue scans a value into the destination.
func scanValue(dest, value any) error {
	if !canScan(dest, value) {
		return nil
	}

	// Handle pgtype types first
	result := scanAndAssign(dest, value)
	if result == assignUUID || result == assignPGType {
		return nil
	}

	// Handle simple types (string, int, etc.) via reflection
	rv := reflect.ValueOf(dest)
	elem := rv.Elem()
	if !elem.CanSet() {
		return nil
	}

	val := reflect.ValueOf(value)
	switch {
	case val.Type().AssignableTo(elem.Type()):
		elem.Set(val)
	case val.Type().ConvertibleTo(elem.Type()):
		elem.Set(val.Convert(elem.Type()))
	case val.Kind() == reflect.Struct && elem.Kind() == reflect.Struct:
		copyStructFields(elem, val)
	}

	return nil
}

type assignResult uint8

const (
	assignNone assignResult = iota
	assignUUID
	assignPGType
	assignSimple
)

// canScan checks if scanning is possible.
func canScan(dest, value any) bool {
	if dest == nil {
		return false
	}

	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false
	}

	if value == nil {
		return false
	}

	return true
}

// scanAndAssign handles type-specific scanning and returns the result type.
func scanAndAssign(dest, value any) assignResult {
	switch v := dest.(type) {
	case *pgtype.UUID:
		return scanUUID(v, value)
	case *pgtype.Text:
		return scanText(v, value)
	case *pgtype.Bool:
		return scanBool(v, value)
	case *pgtype.Numeric:
		return scanNumeric(v, value)
	case *pgtype.Timestamptz:
		return scanTimestamptz(v, value)
	}

	return assignSimple
}

// scanUUID handles UUID type scanning.
func scanUUID(dest *pgtype.UUID, value any) assignResult {
	switch v := value.(type) {
	case pgtype.UUID:
		*dest = v
		return assignUUID
	case uuid.UUID:
		*dest = pgtype.UUID{Bytes: v, Valid: true}
		return assignUUID
	}
	_ = dest.Scan(value)
	return assignUUID
}

// scanText handles Text type scanning.
func scanText(dest *pgtype.Text, value any) assignResult {
	switch v := value.(type) {
	case pgtype.Text:
		*dest = v
	case string:
		dest.String = v
		dest.Valid = true
	}
	return assignPGType
}

// scanBool handles Bool type scanning.
func scanBool(dest *pgtype.Bool, value any) assignResult {
	switch v := value.(type) {
	case pgtype.Bool:
		*dest = v
	case bool:
		dest.Bool = v
		dest.Valid = true
	}
	return assignPGType
}

// scanNumeric handles Numeric type scanning.
func scanNumeric(dest *pgtype.Numeric, value any) assignResult {
	if v, ok := value.(pgtype.Numeric); ok {
		*dest = v
	}
	return assignPGType
}

// scanTimestamptz handles Timestamptz type scanning.
func scanTimestamptz(dest *pgtype.Timestamptz, value any) assignResult {
	switch v := value.(type) {
	case pgtype.Timestamptz:
		*dest = v
	case time.Time:
		dest.Time = v
		dest.Valid = true
	}
	return assignPGType
}

// copyStructFields copies fields from src struct to dest struct by name.
func copyStructFields(dest, src reflect.Value) {
	destType := dest.Type()
	srcType := src.Type()

	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		srcField, ok := srcType.FieldByName(destField.Name)
		if !ok {
			continue
		}

		destFieldValue := dest.Field(i)
		srcFieldValue := src.Field(srcField.Index[0])

		if destFieldValue.CanSet() && srcFieldValue.Type().AssignableTo(destFieldValue.Type()) {
			destFieldValue.Set(srcFieldValue)
		}
	}
}

// Helper functions for creating pgtype values in tests

// NewUUID creates a valid pgtype.UUID from a uuid.UUID.
func NewUUID(t *testing.T, id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// NewPGTypeUUID creates a pgtype.UUID that works correctly with the mock scanner.
// This is the preferred way to create UUID values for mock rows.
func NewPGTypeUUID(t *testing.T, id uuid.UUID) pgtype.UUID {
	t.Helper()
	return pgtype.UUID{Bytes: id, Valid: true}
}

// NewUUIDFromString creates a valid pgtype.UUID from a string UUID.
func NewUUIDFromString(t *testing.T, idStr string) pgtype.UUID {
	id, err := uuid.Parse(idStr)
	if err != nil {
		t.Fatalf("Failed to parse UUID: %v", err)
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// NewTimestamptz creates a valid pgtype.Timestamptz from a time.Time.
func NewTimestamptz(t *testing.T, tm time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: tm, Valid: true}
}

// NewText creates a valid pgtype.Text from a string.
func NewText(t *testing.T, s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// NewNumeric creates a pgtype.Numeric from a float64.
func NewNumeric(t *testing.T, f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

// NewBool creates a valid pgtype.Bool from a bool.
func NewBool(t *testing.T, b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
