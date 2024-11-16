package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type PersonService struct{}

func NewPersonService() *PersonService {
	return &PersonService{}
}

type PaginationResultPerson struct {
	Total       int64           `json:"total"`
	CurrentPage int             `json:"current_page"`
	PageSize    int             `json:"page_size"`
	People      []models.Person `json:"data"`
}

func (s *PersonService) GetPeople(currentPage, pageSize int, search string) (*PaginationResultPerson, error) {
	var people []models.Person
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Person{})

	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Limit(pageSize).Offset(offset).Find(&people).Error
	if err != nil {
		return nil, err
	}

	return &PaginationResultPerson{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		People:      people,
	}, nil
}

func (s *PersonService) CreatePerson(person models.Person) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "person", "id", "person_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	return config.DB.Create(&person).Error
}

func (s *PersonService) GetPersonByID(id uuid.UUID) (models.Person, error) {
	var person models.Person
	err := config.DB.First(&person, "id = ?", id).Error
	return person, err
}

func (s *PersonService) UpdatePerson(person models.Person) error {
	return config.DB.Save(&person).Error
}

func (s *PersonService) DeletePerson(id uuid.UUID) error {
	return config.DB.Delete(&models.Person{}, id).Error
}
