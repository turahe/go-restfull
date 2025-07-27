package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"webapi/internal/db/model"
	"webapi/internal/interfaces/http/controllers"
	"webapi/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaxonomyRepo struct {
	mock.Mock
	repository.TaxonomyRepository
}

func (m *MockTaxonomyRepo) Create(ctx context.Context, taxonomy *model.Taxonomy) error {
	args := m.Called(ctx, taxonomy)
	return args.Error(0)
}
func (m *MockTaxonomyRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Taxonomy, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Taxonomy), args.Error(1)
}
func (m *MockTaxonomyRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Taxonomy, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Taxonomy), args.Error(1)
}
func (m *MockTaxonomyRepo) Update(ctx context.Context, taxonomy *model.Taxonomy) error {
	args := m.Called(ctx, taxonomy)
	return args.Error(0)
}
func (m *MockTaxonomyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockTaxonomyRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockTaxonomyRepo) Search(ctx context.Context, query string, limit, offset int) ([]*model.Taxonomy, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*model.Taxonomy), args.Error(1)
}

func TestCreateTaxonomy_Success(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockTaxonomyRepo)
	h := controllers.NewTaxonomyHandler(mockRepo)
	app.Post("/taxonomies", h.CreateTaxonomy)

	tax := &model.Taxonomy{ID: uuid.New().String(), Name: "Test", Description: "desc"}
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(tax)
	req := httptest.NewRequest(http.MethodPost, "/taxonomies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestGetTaxonomyByID_NotFound(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockTaxonomyRepo)
	h := controllers.NewTaxonomyHandler(mockRepo)
	app.Get("/taxonomies/:id", h.GetTaxonomyByID)

	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(&model.Taxonomy{}, errors.New("not found"))

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/taxonomies/"+id, nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetAllTaxonomies_Success(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockTaxonomyRepo)
	h := controllers.NewTaxonomyHandler(mockRepo)
	app.Get("/taxonomies", h.GetAllTaxonomies)

	mockRepo.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return([]*model.Taxonomy{}, nil)
	mockRepo.On("Count", mock.Anything).Return(int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/taxonomies", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdateTaxonomy_Success(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockTaxonomyRepo)
	h := controllers.NewTaxonomyHandler(mockRepo)
	app.Put("/taxonomies/:id", h.UpdateTaxonomy)

	existingTax := &model.Taxonomy{ID: uuid.New().String(), Name: "Existing", Description: "existing"}
	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(existingTax, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	id := uuid.New().String()
	updateData := map[string]string{"name": "Updated", "description": "updated"}
	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest(http.MethodPut, "/taxonomies/"+id, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteTaxonomy_Success(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockTaxonomyRepo)
	h := controllers.NewTaxonomyHandler(mockRepo)
	app.Delete("/taxonomies/:id", h.DeleteTaxonomy)

	mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/taxonomies/"+id, nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
} 