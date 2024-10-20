package domain

// Api wrapper for prospect

type (
	Rating struct {
		// rating
		Rate  float64 `json:"rate"`
		Count int     `json:"count"`
	} // @name Rating

	Product struct {
		Id          int     `json:"id"`
		Title       string  `json:"title"`
		Price       float64 `json:"price"`
		Description string  `json:"description"`
		Image       string  `json:"image"`
	} // @name  Product

)

type (
	ProcpectaService interface {
		// GetProduct by category
		GetProductByCategory(category string) (result []Product, err error)
		// CreateProduct
		CreateProduct(product Product) (result Product, err error)
	} // @name  ProcpectaService

)

const (
	ApiUrl = "https://fakestoreapi.com/"
)
