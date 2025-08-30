package models

// Category represents a product category.
// It includes a unique code and a name.
type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"unique;not null"`
	Name string `gorm:"not null"`
}

func (Category) TableName() string {
	return "categories"
}
