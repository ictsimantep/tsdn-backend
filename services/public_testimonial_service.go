package services

import (
	"backend-school/config"
	"backend-school/models"

	"github.com/google/uuid"
)

type TestimonialService struct{}

func NewTestimonialService() *TestimonialService {
	return &TestimonialService{}
}

type PaginationResultTestimonial struct {
	Total        int64                `json:"total"`
	CurrentPage  int                  `json:"current_page"`
	PageSize     int                  `json:"page_size"`
	Testimonials []models.Testimonial `json:"data"`
}

func (s *TestimonialService) GetTestimonials(currentPage, pageSize int, search string) (*PaginationResultTestimonial, error) {
	var testimonials []models.Testimonial
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Testimonial{})

	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Limit(pageSize).Offset(offset).Find(&testimonials).Error
	if err != nil {
		return nil, err
	}

	return &PaginationResultTestimonial{
		Total:        total,
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		Testimonials: testimonials,
	}, nil
}

func (s *TestimonialService) CreateTestimonial(testimonial models.Testimonial) error {
	return config.DB.Create(&testimonial).Error
}

func (s *TestimonialService) GetTestimonialByID(id uuid.UUID) (models.Testimonial, error) {
	var testimonial models.Testimonial
	err := config.DB.First(&testimonial, "id = ?", id).Error
	return testimonial, err
}

func (s *TestimonialService) UpdateTestimonial(testimonial models.Testimonial) error {
	return config.DB.Save(&testimonial).Error
}

func (s *TestimonialService) DeleteTestimonial(id uuid.UUID) error {
	return config.DB.Delete(&models.Testimonial{}, id).Error
}
