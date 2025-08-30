package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	defaultLimit  = 10
	maxLimit      = 100
	defaultOffset = 0
)

// ProductProvider defines the interface for fetching product and category data.
// This allows for dependency inversion, making the handler testable.
type ProductProvider interface {
	GetProducts(params models.GetProductsParams) ([]models.Product, int64, error)
	GetProductByCode(code string) (*models.Product, error)
	GetAllCategories() ([]models.Category, error)
	CreateCategory(category *models.Category) error
}

// Response structs for building the JSON output.
type VariantResponse struct {
	Name  string          `json:"name"`
	SKU   string          `json:"sku"`
	Price decimal.Decimal `json:"price"`
}

type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductResponse struct {
	Code     string            `json:"code"`
	Price    decimal.Decimal   `json:"price"`
	Category CategoryResponse  `json:"category"`
	Variants []VariantResponse `json:"variants"`
}

type CatalogResponse struct {
	Total    int64             `json:"total"`
	Products []ProductResponse `json:"products"`
}

// CatalogHandler handles HTTP requests for the catalog.
type CatalogHandler struct {
	repo ProductProvider
}

// NewCatalogHandler creates a new instance of CatalogHandler.
func NewCatalogHandler(r ProductProvider) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

// HandleGet handles the GET /catalog endpoint.
func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = defaultOffset
	}

	var priceLessThan *decimal.Decimal
	if p := r.URL.Query().Get("priceLessThan"); p != "" {
		if val, err := decimal.NewFromString(p); err == nil {
			priceLessThan = &val
		}
	}

	categoryCode := r.URL.Query().Get("category")

	params := models.GetProductsParams{
		Offset:        offset,
		Limit:         limit,
		PriceLessThan: priceLessThan,
		CategoryCode:  categoryCode,
	}
	products, total, err := h.repo.GetProducts(params)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		variantResponses := make([]VariantResponse, len(p.Variants))
		for j, v := range p.Variants {
			variantPrice := v.Price
			if variantPrice.IsZero() {
				variantPrice = p.Price
			}
			variantResponses[j] = VariantResponse{Name: v.Name, SKU: v.SKU, Price: variantPrice}
		}
		productResponses[i] = ProductResponse{
			Code:     p.Code,
			Price:    p.Price,
			Category: CategoryResponse{Code: p.Category.Code, Name: p.Category.Name},
			Variants: variantResponses,
		}
	}

	response := CatalogResponse{
		Total:    total,
		Products: productResponses,
	}

	api.OKResponse(w, response)
}

// HandleGetDetails handles the GET /catalog/:code endpoint.
// It fetches a single product by its code and returns its detailed information.
func (h *CatalogHandler) HandleGetDetails(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	product, err := h.repo.GetProductByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "Product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch product")
		return
	}

	variantResponses := make([]VariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		variantPrice := v.Price
		if variantPrice.IsZero() {
			variantPrice = product.Price
		}
		variantResponses[i] = VariantResponse{Name: v.Name, SKU: v.SKU, Price: variantPrice}
	}

	response := ProductResponse{
		Code:     product.Code,
		Price:    product.Price,
		Category: CategoryResponse{Code: product.Category.Code, Name: product.Category.Name},
		Variants: variantResponses,
	}

	api.OKResponse(w, response)
}

// HandleGetCategories handles the GET /categories endpoint.
// It returns a list of all available product categories.
func (h *CatalogHandler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	categoryResponses := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		categoryResponses[i] = CategoryResponse{
			Code: c.Code,
			Name: c.Name,
		}
	}

	api.OKResponse(w, categoryResponses)
}

// HandleCreateCategory handles the POST /categories endpoint.
// It creates a new product category based on the provided JSON body.
func (h *CatalogHandler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var newCategory models.Category

	if err := json.NewDecoder(r.Body).Decode(&newCategory); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if newCategory.Code == "" || newCategory.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Category code and name are required")
		return
	}

	if err := h.repo.CreateCategory(&newCategory); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Failed to create category")
		return
	}

	response := CategoryResponse{
		Code: newCategory.Code,
		Name: newCategory.Name,
	}

	api.CreatedResponse(w, response)
}
