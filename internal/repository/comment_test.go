package repository

import (
	"context"
	"testing"
	"webapi/internal/db/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Minimal DB interface for mocking
// (In production, use pgxpool.Pool which implements this)
type DBIface interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgx.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error) {
	argsMock := m.Called(ctx, sql, args)
	return pgx.CommandTag("INSERT 0 1"), argsMock.Error(1)
}
func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgx.Row)
}
func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	argsMock := m.Called(ctx, sql, args)
	return nil, argsMock.Error(1)
}

func TestInsertNestedComment_Mock(t *testing.T) {
	mockDB := new(MockDB)
	repo := &CommentRepositoryImpl{db: mockDB}
	parentID := uuid.New()
	comment := &model.Comment{ID: uuid.New()}

	// Mock parent right/depth
	mockRow := new(mockRow)
	mockRow.On("Scan", mock.AnythingOfType("*int64"), mock.AnythingOfType("*int64")).Return(nil).Run(func(args mock.Arguments) {
		*(args[0].(*int64)) = 10
		*(args[1].(*int64)) = 1
	})
	mockDB.On("QueryRow", mock.Anything, mock.MatchedBy(func(sql string) bool { return true }), mock.Anything).Return(mockRow)
	mockDB.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool { return true }), mock.Anything).Return(pgx.CommandTag("UPDATE 1"), nil)
	mockDB.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool { return true }), mock.Anything).Return(pgx.CommandTag("INSERT 0 1"), nil)

	err := repo.InsertNestedComment(context.Background(), parentID, comment)
	assert.NoError(t, err)
}

type mockRow struct {
	mock.Mock
}

func (m *mockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *mockRow) FieldDescriptions() []pgx.FieldDescription { return nil }
func (m *mockRow) Values() ([]interface{}, error)            { return nil, nil }
func (m *mockRow) Err() error                                { return nil }

// NOTE: These are illustrative tests. For real tests, use a test Postgres DB and set up/tear down data.

func TestInsertNestedComment_NoPanic(t *testing.T) {
	// This test checks that the method can be called without panicking.
	// In a real test, use a test DB and assert on the DB state.
	repo := &CommentRepositoryImpl{pgxPool: nil, redisClient: nil}
	parentID := uuid.New()
	comment := &model.Comment{ID: uuid.New()}
	// Should not panic (will error due to nil pool)
	err := repo.InsertNestedComment(context.Background(), parentID, comment)
	assert.Error(t, err)
}

func TestGetDescendants_NoPanic(t *testing.T) {
	repo := &CommentRepositoryImpl{pgxPool: nil, redisClient: nil}
	id := uuid.New()
	_, err := repo.GetDescendants(context.Background(), id)
	assert.Error(t, err)
}

func TestGetCommentTree_NoPanic(t *testing.T) {
	repo := &CommentRepositoryImpl{pgxPool: nil, redisClient: nil}
	id := uuid.New()
	_, err := repo.GetCommentTree(context.Background(), id)
	assert.Error(t, err)
}
