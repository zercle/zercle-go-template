package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestDBTXMock_DefaultExec(t *testing.T) {
	mock := &DBTXMock{}

	result, err := mock.Exec(context.Background(), "DELETE FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDBTXMock_DefaultQuery(t *testing.T) {
	mock := &DBTXMock{}

	rows, err := mock.Query(context.Background(), "SELECT * FROM users")
	assert.NoError(t, err)
	assert.Nil(t, rows) // Returns nil when no QueryFunc set
	if rows != nil {
		rows.Close()
	}
}

func TestDBTXMock_CustomQuery(t *testing.T) {
	mock := &DBTXMock{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
			return NewRowsMock([][]interface{}{
				{NewUUID(t, uuid.New()), "test@example.com"},
			}), nil
		},
	}

	rows, err := mock.Query(context.Background(), "SELECT * FROM users")
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	rows.Close()
}

func TestDBTXMock_CustomQueryRow(t *testing.T) {
	mock := &DBTXMock{
		QueryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return NewRowMock([]interface{}{
				NewUUID(t, uuid.New()),
				"test@example.com",
			})
		},
	}

	row := mock.QueryRow(context.Background(), "SELECT * FROM users WHERE id = $1", uuid.New())
	assert.NotNil(t, row)
}

func TestRowsMock_Scan(t *testing.T) {
	rows := [][]interface{}{
		{NewUUID(t, uuid.New()), "test@example.com", NewText(t, "Test User")},
	}
	rowsMock := NewRowsMock(rows)

	var id pgtype.UUID
	var email string
	var name pgtype.Text

	assert.True(t, rowsMock.Next())
	err := rowsMock.Scan(&id, &email, &name)
	assert.NoError(t, err)
	assert.True(t, id.Valid)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "Test User", name.String)
}

func TestRowsMock_EmptyResult(t *testing.T) {
	rowsMock := NewRowsMock([][]interface{}{})

	assert.False(t, rowsMock.Next())
	values, _ := rowsMock.Values()
	assert.Nil(t, values)
	assert.Nil(t, rowsMock.RawValues())
}

func TestRowsMock_CommandTag(t *testing.T) {
	rowsMock := NewRowsMock([][]interface{}{
		{NewUUID(t, uuid.New())},
		{NewUUID(t, uuid.New())},
	})

	tag := rowsMock.CommandTag()
	assert.NotNil(t, tag)
}

func TestRowsMock_FieldDescriptions(t *testing.T) {
	rowsMock := NewRowsMock([][]interface{}{})

	fields := rowsMock.FieldDescriptions()
	assert.NotNil(t, fields)
	assert.Empty(t, fields)
}

func TestRowMock_Scan(t *testing.T) {
	row := NewRowMock([]interface{}{
		NewUUID(t, uuid.New()),
		"test@example.com",
	})

	var id pgtype.UUID
	var email string

	err := row.Scan(&id, &email)
	assert.NoError(t, err)
	assert.True(t, id.Valid)
	assert.Equal(t, "test@example.com", email)
}

func TestRowMock_ScanWithError(t *testing.T) {
	row := &RowMock{
		row: []interface{}{},
	}

	err := row.Scan()
	// Empty row returns pgx.ErrNoRows
	assert.Error(t, err)
	assert.Equal(t, pgx.ErrNoRows, err)
}

func TestNewPGTypeUUID(t *testing.T) {
	id := uuid.New()
	pgUUID := NewPGTypeUUID(t, id)

	assert.True(t, pgUUID.Valid)
	// uuid.UUID is a [16]byte array, pgtype.UUID stores it in Bytes field
	assert.Equal(t, uuid.UUID(pgUUID.Bytes), id)
}

func TestNewUUIDFromString(t *testing.T) {
	idStr := "550e8400-e29b-41d4-a716-446655440000"
	pgUUID := NewUUIDFromString(t, idStr)

	assert.True(t, pgUUID.Valid)
}

func TestNewTimestamptz(t *testing.T) {
	testTime := parseTime(t, "2024-01-15T10:30:00Z")
	ts := NewTimestamptz(t, testTime)

	assert.True(t, ts.Valid)
	assert.Equal(t, testTime, ts.Time)
}

func TestNewText(t *testing.T) {
	text := NewText(t, "Hello World")
	assert.True(t, text.Valid)
	assert.Equal(t, "Hello World", text.String)

	emptyText := NewText(t, "")
	assert.False(t, emptyText.Valid)
}

func TestNewNumeric(t *testing.T) {
	numeric := NewNumeric(t, 100.50)
	assert.NotNil(t, numeric)
}

func TestNewBool(t *testing.T) {
	boolTrue := NewBool(t, true)
	assert.True(t, boolTrue.Bool)
	assert.True(t, boolTrue.Valid)

	boolFalse := NewBool(t, false)
	assert.False(t, boolFalse.Bool)
	assert.True(t, boolFalse.Valid)
}

// Helper function for parsing time in tests
func parseTime(t *testing.T, timeStr string) (result time.Time) {
	t.Helper()
	var err error
	result, err = time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Fatalf("Failed to parse time: %v", err)
	}
	return result
}
