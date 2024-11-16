package services

import (
	"backend-school/config"
	"backend-school/models"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

type CategoryDocumentService struct{}

// Initialize CategoryDocumentService
func NewCategoryDocumentService() *CategoryDocumentService {
	return &CategoryDocumentService{}
}

// RoleHasRulePayload defines the structure for each role rule entry in the request payload
type RoleHasRulePayload struct {
	RoleGuardName string `json:"role_guard_name"`
	Action        string `json:"action"`
}

// CategoryDocumentPayload defines the structure for the create request payload
type CategoryDocumentPayload struct {
	Name         string               `json:"name"`
	Prefix       string               `json:"prefix"`
	RoleHasRules []RoleHasRulePayload `json:"role_has_rules"`
}

func isDuplicateKeyError(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == "23505" // 23505 is the PostgreSQL error code for unique constraint violation
	}
	return false
}

// GetCategoryDocumentByUUID fetches a CategoryDocument and related RoleHasRule entries by UUID
func (s *CategoryDocumentService) GetCategoryDocumentByUUID(uuid string) (map[string]interface{}, error) {
	var categoryDocument models.CategoryDocument
	var roleHasRules []models.RoleHasRule

	// Find the category document
	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at IS NULL").First(&categoryDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category document not found")
		}
		return nil, err
	}

	// Fetch related role_has_rule entries using prefix from category_document
	if err := config.DB.
		Where("category = ?", categoryDocument.Prefix).
		Find(&roleHasRules).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch related role rules: %w", err)
	}

	// Build the response map with category document and related role rules
	result := map[string]interface{}{
		"category_document": categoryDocument,
		"role_has_rules":    roleHasRules,
	}

	return result, nil
}

func (s *CategoryDocumentService) GetCategoryDocumentsPaginated(currentPage, pageSize int, search string, showAll string) (map[string]interface{}, error) {
	var categoryDocuments []models.CategoryDocument
	var totalRecords int64

	offset := (currentPage - 1) * pageSize
	query := config.DB.Model(&models.CategoryDocument{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to count category documents: %w", err)
	}

	if showAll == "" {
		if err := query.Offset(offset).Limit(pageSize).Find(&categoryDocuments).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch category documents: %w", err)
		}
	} else {
		if err := query.Find(&categoryDocuments).Error; err != nil {
			return nil, errors.New("failed to fetch document types")
		}
	}

	// Create a response structure to hold each category document with its associated role_has_rules
	var categoryDocumentsWithRoles []map[string]interface{}
	for _, categoryDocument := range categoryDocuments {
		// Retrieve only role_guard_name and action fields for related role_has_rule entries
		var roleRules []struct {
			RoleGuardName string `json:"role_guard_name"`
			Action        string `json:"action"`
		}
		if err := config.DB.Model(&models.RoleHasRule{}).
			Select("role_guard_name, action").
			Where("category = ?", categoryDocument.Prefix).
			Scan(&roleRules).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch related role rules: %w", err)
		}

		// Add the category document and its role rules to the result
		categoryDocumentWithRoles := map[string]interface{}{
			"category_document": categoryDocument,
			"role_has_rules":    roleRules,
		}
		categoryDocumentsWithRoles = append(categoryDocumentsWithRoles, categoryDocumentWithRoles)
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	result := map[string]interface{}{
		"data":          categoryDocumentsWithRoles,
		"current_page":  currentPage,
		"per_page":      pageSize,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

func (s *CategoryDocumentService) AddCategoryDocument(payload *CategoryDocumentPayload) (*models.CategoryDocument, error) {
	var categoryDocument models.CategoryDocument

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Check if Prefix already exists in category_document
		var existingCategoryDocument models.CategoryDocument
		if err := tx.Where("prefix = ?", payload.Prefix).First(&existingCategoryDocument).Error; err == nil {
			return fmt.Errorf("category document with prefix '%s' already exists", payload.Prefix)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check for existing category document: %w", err)
		}

		// Create the CategoryDocument entry
		categoryDocument = models.CategoryDocument{
			UUID:   uuid.New(),
			Name:   payload.Name,
			Prefix: payload.Prefix,
		}

		if err := tx.Create(&categoryDocument).Error; err != nil {
			return fmt.Errorf("failed to create category document: %w", err)
		}

		// Step 2: Create entries in role_has_rule table, ensuring no duplicates
		for _, rule := range payload.RoleHasRules {
			var existingRoleHasRule models.RoleHasRule
			if err := tx.Where("role_guard_name = ? AND action = ? AND category = ?", rule.RoleGuardName, rule.Action, payload.Prefix).First(&existingRoleHasRule).Error; err == nil {
				// If a duplicate is found, skip adding this rule
				continue
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to check for existing role has rule: %w", err)
			}

			// Create new role_has_rule if no duplicate found
			roleHasRule := models.RoleHasRule{
				RoleGuardName: rule.RoleGuardName,
				RulePolicy:    "document",
				Action:        rule.Action,
				Category:      payload.Prefix,
			}
			if err := tx.Create(&roleHasRule).Error; err != nil {
				return fmt.Errorf("failed to create role has rule: %w", err)
			}
		}

		// Step 3: Create entries in casbin_rule table, ensuring no duplicates
		for _, rule := range payload.RoleHasRules {
			var existingCasbinRule models.CasbinRule
			casbinRule := models.CasbinRule{
				Ptype: "p",
				V0:    rule.RoleGuardName,
				V1:    "document",
				V2:    rule.Action,
				V3:    payload.Prefix,
			}

			// Check for duplicate in casbin_rule
			if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ?", casbinRule.Ptype, casbinRule.V0, casbinRule.V1, casbinRule.V2, casbinRule.V3).First(&existingCasbinRule).Error; err == nil {
				// If duplicate found, skip adding this rule
				continue
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to check for existing casbin rule: %w", err)
			}

			// Create new casbin_rule if no duplicate found
			if err := tx.Create(&casbinRule).Error; err != nil {
				return fmt.Errorf("failed to create casbin rule: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return &categoryDocument, nil
}

func (s *CategoryDocumentService) UpdateCategoryDocumentByUUID(uuid string, payload *CategoryDocumentPayload) (*models.CategoryDocument, error) {
	var categoryDocument models.CategoryDocument

	// Find the category document by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&categoryDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("category document not found")
		}
		return nil, fmt.Errorf("failed to find category document: %w", err)
	}

	// Update name and prefix to match payload
	categoryDocument.Name = payload.Name
	categoryDocument.Prefix = payload.Prefix
	categoryDocument.UpdatedAt = time.Now()

	// Start transaction to handle role updates
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Update the category document's name and prefix
		if err := tx.Save(&categoryDocument).Error; err != nil {
			return fmt.Errorf("failed to update category document: %w", err)
		}

		// Fetch existing RoleHasRule entries for this document's prefix
		var existingRules []models.RoleHasRule
		if err := tx.Where("category = ?", categoryDocument.Prefix).Find(&existingRules).Error; err != nil {
			return fmt.Errorf("failed to fetch existing role rules: %w", err)
		}

		// Prepare maps for quick lookup
		existingRulesMap := make(map[string]models.RoleHasRule)
		for _, rule := range existingRules {
			key := fmt.Sprintf("%s:%s", rule.RoleGuardName, rule.Action)
			existingRulesMap[key] = rule
		}

		// Process role_has_rules from payload
		for _, rule := range payload.RoleHasRules {
			key := fmt.Sprintf("%s:%s", rule.RoleGuardName, rule.Action)

			if existingRule, found := existingRulesMap[key]; found {
				// Update existing rule if found
				existingRule.Category = payload.Prefix
				existingRule.RulePolicy = "document"
				if err := tx.Save(&existingRule).Error; err != nil {
					return fmt.Errorf("failed to update role rule: %w", err)
				}
				// Remove from map after processing to track what's left
				delete(existingRulesMap, key)
			} else {
				// Create new rule if not found
				newRule := models.RoleHasRule{
					RoleGuardName: rule.RoleGuardName,
					RulePolicy:    "document",
					Action:        rule.Action,
					Category:      payload.Prefix,
				}
				if err := tx.Create(&newRule).Error; err != nil {
					return fmt.Errorf("failed to create new role rule: %w", err)
				}
			}
		}

		// Delete any remaining rules in existingRulesMap that are not in the payload
		for _, rule := range existingRulesMap {
			if err := tx.Delete(&rule).Error; err != nil {
				return fmt.Errorf("failed to delete obsolete role rule: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return &categoryDocument, nil
}

// DeleteCategoryDocument soft deletes a category document by setting the deleted_at timestamp
func (s *CategoryDocumentService) DeleteCategoryDocument(uuid string) error {
	currentTime := time.Now()

	// Check if category document with UUID exists
	var categoryDocument models.CategoryDocument
	if err := config.DB.Where("uuid = ?", uuid).First(&categoryDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("category document not found")
		}
		return errors.New("failed to check category document")
	}

	// Soft delete by setting deleted_at timestamp
	if err := config.DB.Model(&categoryDocument).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete category document")
	}

	return nil
}
