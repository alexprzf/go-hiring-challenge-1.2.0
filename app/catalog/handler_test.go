package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockProductProvider struct {
	GetProductsFunc      func(params models.GetProductsParams) ([]models.Product, int64, error)
	GetProductByCodeFunc func(code string) (*models.Product, error)
	GetAllCategoriesFunc func() ([]models.Category, error)
	CreateCategoryFunc   func(category *models.Category) error
}

func (m *mockProductProvider) GetProducts(params models.GetProductsParams) ([]models.Product, int64, error) {
	return m.GetProductsFunc(params)
}
func (m *mockProductProvider) GetProductByCode(code string) (*models.Product, error) {
	return m.GetProductByCodeFunc(code)
}
func (m *mockProductProvider) GetAllCategories() ([]models.Category, error) {
	return m.GetAllCategoriesFunc()
}
func (m *mockProductProvider) CreateCategory(category *models.Category) error {
	return m.CreateCategoryFunc(category)
}

func TestHandleGet(t *testing.T) {
	mockRepo := &mockProductProvider{
		GetProductsFunc: func(params models.GetProductsParams) ([]models.Product, int64, error) {
			assert.Equal(t, 5, params.Limit)
			assert.Equal(t, 0, params.Offset)

			products := []models.Product{{Code: "P001", Price: decimal.NewFromInt(100)}}
			return products, 1, nil
		},
	}
	handler := NewCatalogHandler(mockRepo)
	req := httptest.NewRequest("GET", "/catalog?limit=5", nil)
	rr := httptest.NewRecorder()

	handler.HandleGet(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response CatalogResponse
	json.NewDecoder(rr.Body).Decode(&response)
	assert.Equal(t, int64(1), response.Total)
	assert.Len(t, response.Products, 1)
}

func TestHandleGetDetails(t *testing.T) {
	mockRepo := &mockProductProvider{
		GetProductByCodeFunc: func(code string) (*models.Product, error) {
			if code == "P001" {
				return &models.Product{Code: "P001"}, nil
			}
			return nil, gorm.ErrRecordNotFound
		},
	}
	handler := NewCatalogHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/catalog/P001", nil)
		req.SetPathValue("code", "P001")
		rr := httptest.NewRecorder()
		handler.HandleGetDetails(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response ProductResponse
		json.NewDecoder(rr.Body).Decode(&response)
		assert.Equal(t, "P001", response.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/catalog/P999", nil)
		req.SetPathValue("code", "P999")
		rr := httptest.NewRecorder()
		handler.HandleGetDetails(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestHandleGetCategories(t *testing.T) {
	mockRepo := &mockProductProvider{
		GetAllCategoriesFunc: func() ([]models.Category, error) {
			return []models.Category{{Code: "cat1", Name: "Category 1"}}, nil
		},
	}
	handler := NewCatalogHandler(mockRepo)
	req := httptest.NewRequest("GET", "/categories", nil)
	rr := httptest.NewRecorder()

	handler.HandleGetCategories(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response []CategoryResponse
	json.NewDecoder(rr.Body).Decode(&response)
	assert.Len(t, response, 1)
	assert.Equal(t, "cat1", response[0].Code)
}

func TestHandleCreateCategory(t *testing.T) {
	mockRepo := &mockProductProvider{
		CreateCategoryFunc: func(category *models.Category) error {
			assert.Equal(t, "new-cat", category.Code)
			assert.Equal(t, "New Category", category.Name)
			return nil
		},
	}
	handler := NewCatalogHandler(mockRepo)

	t.Run("Success", func(t *testing.T) {
		body := `{"code": "new-cat", "name": "New Category"}`
		req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
		rr := httptest.NewRecorder()
		handler.HandleCreateCategory(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		body := `{"code": "bad"`
		req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
		rr := httptest.NewRecorder()
		handler.HandleCreateCategory(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
