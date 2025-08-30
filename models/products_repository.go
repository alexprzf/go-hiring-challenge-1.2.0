package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

type GetProductsParams struct {
	Offset        int
	Limit         int
	PriceLessThan *decimal.Decimal
	CategoryCode  string
}

func (r *ProductsRepository) GetProducts(params GetProductsParams) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{})

	if params.CategoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", params.CategoryCode)
	}

	if params.PriceLessThan != nil {
		query = query.Where("products.price < ?", *params.PriceLessThan)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(params.Offset).Limit(params.Limit).
		Preload("Variants").
		Preload("Category").
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	err := r.db.
		Preload("Variants").
		Preload("Category").
		Where("code = ?", code).
		First(&product).Error

	if err != nil {
		return nil, err
	}
	return &product, nil
}
