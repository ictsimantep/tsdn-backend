// services/page_service.go
package services

import (
	"backend-school/config"
	"backend-school/models"
)

// PageService handles all page-related logic
type PageService struct {
}

// NewPageService creates a new instance of PageService
func NewPageService() *PageService {
	return &PageService{}
}

// GetAllPages retrieves all pages from the database
func (s *PageService) GetAllPages() ([]models.Page, error) {

	var pages []models.Page
	err := config.DB.Find(&pages).Error
	return pages, err
}

// GetPageBySlug retrieves a single page by its slug
func (s *PageService) GetPageBySlug(slug string) (*models.Page, error) {
	var page models.Page
	err := config.DB.Where("slug = ?", slug).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}
