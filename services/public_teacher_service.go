package services

import (
	"backend-school/config"
	"backend-school/models"

	"github.com/google/uuid"
)

type TeacherService struct{}

func NewTeacherService() *TeacherService {
	return &TeacherService{}
}

type PaginationResultTeacher struct {
	Total       int64            `json:"total"`
	CurrentPage int              `json:"current_page"`
	PageSize    int              `json:"page_size"`
	Teachers    []models.Teacher `json:"data"`
}

func (s *TeacherService) GetTeachers(currentPage, pageSize int, search string) (*PaginationResultTeacher, error) {
	var teachers []models.Teacher
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Teacher{})

	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Limit(pageSize).Offset(offset).Find(&teachers).Error
	if err != nil {
		return nil, err
	}

	return &PaginationResultTeacher{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		Teachers:    teachers,
	}, nil
}

func (s *TeacherService) CreateTeacher(teacher models.Teacher) error {
	return config.DB.Create(&teacher).Error
}

func (s *TeacherService) GetTeacherByID(id uuid.UUID) (models.Teacher, error) {
	var teacher models.Teacher
	err := config.DB.First(&teacher, "id = ?", id).Error
	return teacher, err
}

func (s *TeacherService) UpdateTeacher(teacher models.Teacher) error {
	return config.DB.Save(&teacher).Error
}

func (s *TeacherService) DeleteTeacher(id uuid.UUID) error {
	return config.DB.Delete(&models.Teacher{}, id).Error
}
