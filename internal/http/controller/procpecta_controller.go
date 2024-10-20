package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/domain"
	"github.com/chatApp/internal/http/transport"
)

type ProspectaContoller struct {
	ps domain.ProcpectaService
}

func NewProspectaController(ps domain.ProcpectaService) ProspectaContoller {
	return ProspectaContoller{
		ps: ps,
	}
}

// GetProduct retrieves products based on the category.
//
//	@Summary		Get Products by Category
//	@Description	Retrieve a list of products that belong to a specific category.
//	@Tags			Product
//	@ID				getProductByCategory
//	@Accept			json
//	@Produce		json
//	@Param			cat	path		string	true	"Product Category"
//	@Success		200	{object}	domain.BaseResponse{data=[]domain.Product}
//	@Failure		400	{object}	domain.InvalidRequestError
//	@Failure		404	{object}	domain.NotFoundError
//	@Failure		500	{object}	domain.SystemError
//	@Router			/products/category/{cat} [get]
func (c ProspectaContoller) GetProduct(ctx echo.Context) error {
	cat := ctx.Param("cat")
	products, err := c.ps.GetProductByCategory(cat)
	if err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusOK, products)

}

// CreateProduct adds a new product.
//
//	@Summary		Create a new Product
//	@Description	Add a new product to the system.
//	@Tags			Product
//	@ID				createProduct
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.Product	true	"Product details"
//	@Success		201		{object}	domain.BaseResponse{data=domain.Product}
//	@Failure		400		{object}	domain.InvalidRequestError
//	@Failure		500		{object}	domain.SystemError
//	@Router			/products [post]
func (c ProspectaContoller) CreateProduct(ctx echo.Context) error {
	var product domain.Product
	transport.DecodeAndValidateRequestBody(ctx, &product)
	result, err := c.ps.CreateProduct(product)
	if err != nil {
		return err
	}
	return transport.SendResponse(ctx, http.StatusCreated, result)

}
