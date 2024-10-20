package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chatApp/internal/domain"
)

type procePectaServiceImpl struct {
	//repo domain.UserRepository
}

func NewProcpectaService() domain.ProcpectaService {
	return &procePectaServiceImpl{}
}

// CreateProduct implements domain.ProcpectaService.
func (p *procePectaServiceImpl) CreateProduct(product domain.Product) (result domain.Product, err error) {
	url := fmt.Sprintf(domain.ApiUrl+"%s", "products")
	jsonnData, err := json.Marshal(product)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonnData))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil

}

// GetProductByCategory implements domain.ProcpectaService.
func (p *procePectaServiceImpl) GetProductByCategory(category string) (result []domain.Product, err error) {
	url := fmt.Sprintf(domain.ApiUrl+"%s/%s/%s", "products", "category", category)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil

}
