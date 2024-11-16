package services

import (
	"backend-school/config"
	"backend-school/models"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentTypeService struct{}

// NewDocumentTypeService initializes a new DocumentTypeService
func NewDocumentTypeService() *DocumentTypeService {
	return &DocumentTypeService{}
}

// defines the structure for the create request payload
type DocumentTypePayload struct {
	Name               string               `json:"name"`
	Prefix             string               `json:"prefix"`
	RoleHasRules       []RoleHasRulePayload `json:"role_has_rules"`
	DocumentCategoryID int                  `json:"document_category_id"`
}

func (s *DocumentTypeService) GetDocumentTypesPaginated(currentPage, pageSize int, search string, showAll string) (map[string]interface{}, error) {
	var documentTypes []models.DocumentType
	var totalRecords int64

	offset := (currentPage - 1) * pageSize

	// Query dasar untuk document type
	query := config.DB.Model(&models.DocumentType{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Hitung total dokumen yang sesuai dengan query
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, errors.New("failed to count document types")
	}

	// Ambil data document types dengan paginasi
	if showAll == "" {
		if err := query.Offset(offset).Limit(pageSize).Find(&documentTypes).Error; err != nil {
			return nil, errors.New("failed to fetch document types")
		}
	} else {
		if err := query.Find(&documentTypes).Error; err != nil {
			return nil, errors.New("failed to fetch document types")
		}
	}

	// Membuat hasil akhir dengan memasukkan role_guard_name dan action saja untuk setiap document type
	documentTypesWithRoles := make([]map[string]interface{}, len(documentTypes))
	for i, docType := range documentTypes {
		var roleHasRules []struct {
			RoleGuardName string `json:"role_guard_name"`
			Action        string `json:"action"`
		}
		if err := config.DB.Model(&models.RoleHasRule{}).
			Select("role_guard_name, action").
			Where("type = ?", docType.Prefix).
			Scan(&roleHasRules).Error; err != nil {
			return nil, errors.New("failed to fetch role has rules")
		}

		// Membuat map untuk menyertakan document type dengan role_has_rule yang difilter
		documentTypesWithRoles[i] = map[string]interface{}{
			"uuid":           docType.UUID,
			"name":           docType.Name,
			"prefix":         docType.Prefix,
			"role_has_rules": roleHasRules,
		}
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	// Menyiapkan hasil dalam bentuk map untuk JSON response
	result := map[string]interface{}{
		"data":          documentTypesWithRoles,
		"current_page":  currentPage,
		"per_page":      pageSize,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

func (s *DocumentTypeService) AddDocumentType(payload *DocumentTypePayload) (*models.DocumentType, error) {
	var documentType models.DocumentType

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Check if Prefix already exists in
		var existingDocumentType models.DocumentType
		if err := tx.Where("prefix = ?", payload.Prefix).First(&existingDocumentType).Error; err == nil {
			return fmt.Errorf("type document with prefix '%s' already exists", payload.Prefix)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check for existing type document: %w", err)
		}

		// Create the  entry
		documentType = models.DocumentType{
			UUID:               uuid.New(),
			Name:               payload.Name,
			Prefix:             payload.Prefix,
			DocumentCategoryID: IntPtr(payload.DocumentCategoryID),
		}

		if err := tx.Create(&documentType).Error; err != nil {
			return fmt.Errorf("failed to create type document: %w", err)
		}

		var documentCategory models.CategoryDocument

		// Query untuk mencari data berdasarkan `status_document_id` dari request
		if err := config.DB.Where("id = ?", documentType.DocumentCategoryID).First(&documentCategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Jika data tidak ditemukan
				return fmt.Errorf("failed to check category document: %w", err)
			}
			// Jika ada error lain saat query
			return fmt.Errorf("failed to check category document: %w", err)
		}

		// Step 2: Create entries in role_has_rule table, ensuring no duplicates
		for _, rule := range payload.RoleHasRules {
			var existingRoleHasRule models.RoleHasRule
			if err := tx.Where("role_guard_name = ? AND action = ? AND type = ? AND category = ?", rule.RoleGuardName, rule.Action, payload.Prefix, documentCategory.Prefix).First(&existingRoleHasRule).Error; err == nil {
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
				Type:          payload.Prefix,
				Category:      documentCategory.Prefix,
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
				V3:    documentCategory.Prefix,
				V4:    payload.Prefix,
			}

			// Check for duplicate in casbin_rule
			if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ?", casbinRule.Ptype, casbinRule.V0, casbinRule.V1, casbinRule.V2, casbinRule.V3, casbinRule.V4).First(&existingCasbinRule).Error; err == nil {
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

	return &documentType, nil
}

// GetDocumentTypeByUUID fetches a DocumentType by UUID
func (s *DocumentTypeService) GetDocumentTypeByUUID(uuid string) (*models.DocumentType, error) {
	var documentType models.DocumentType

	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at IS NULL").First(&documentType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document type not found")
		}
		return nil, err
	}

	return &documentType, nil
}

func (s *DocumentTypeService) UpdateDocumentTypeByUUID(uuid string, payload *DocumentTypePayload) (*models.DocumentType, error) {
	var documentType models.DocumentType

	// Find the document type by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&documentType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("type document not found")
		}
		return nil, fmt.Errorf("failed to find type document: %w", err)
	}

	// Simpan kategori dan prefix lama sebelum diubah
	oldCategoryID := documentType.DocumentCategoryID
	oldPrefix := documentType.Prefix

	// Start transaction to handle updates in the correct order
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Update role_has_rule entries

		// Fetch existing RoleHasRule entries for this document's prefix
		var existingRules []models.RoleHasRule
		if err := tx.Where("type = ?", oldPrefix).Find(&existingRules).Error; err != nil {
			return fmt.Errorf("failed to fetch existing role rules: %w", err)
		}

		// Prepare maps for quick lookup
		existingRulesMap := make(map[string]models.RoleHasRule)
		for _, rule := range existingRules {
			key := fmt.Sprintf("%s:%s", rule.RoleGuardName, rule.Action)
			existingRulesMap[key] = rule
		}

		var oldDocumentCategory models.CategoryDocument

		// Query to find the old category document by ID
		if err := config.DB.Where("id = ?", *oldCategoryID).First(&oldDocumentCategory).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to find old category document: %w", err)
			}
			return fmt.Errorf("failed to check old category document: %w", err)
		}

		// Process role_has_rules from payload and delete old Casbin rules
		for _, rule := range payload.RoleHasRules {
			key := fmt.Sprintf("%s:%s", rule.RoleGuardName, rule.Action)

			if existingRule, found := existingRulesMap[key]; found {
				// Hapus aturan Casbin lama menggunakan kategori dan prefix lama
				if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ?",
					"p", existingRule.RoleGuardName, "document", existingRule.Action, oldDocumentCategory.Prefix, oldPrefix).
					Delete(&models.CasbinRule{}).Error; err != nil {
					return fmt.Errorf("failed to delete old Casbin rule: %w", err)
				}

				// Update existing role_has_rule
				existingRule.Type = payload.Prefix
				existingRule.RulePolicy = "document"
				existingRule.Category = oldDocumentCategory.Prefix
				if err := tx.Save(&existingRule).Error; err != nil {
					return fmt.Errorf("failed to update role rule: %w", err)
				}
				delete(existingRulesMap, key)
			} else {
				// Create a new role_has_rule entry if not found
				newRule := models.RoleHasRule{
					RoleGuardName: rule.RoleGuardName,
					RulePolicy:    "document",
					Action:        rule.Action,
					Type:          payload.Prefix,
					Category:      oldDocumentCategory.Prefix,
				}
				if err := tx.Create(&newRule).Error; err != nil {
					return fmt.Errorf("failed to create new role rule: %w", err)
				}
			}
		}

		// Delete any remaining rules in existingRulesMap that are not in the payload and remove corresponding Casbin rules
		for _, rule := range existingRulesMap {
			if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ?",
				"p", rule.RoleGuardName, "document", rule.Action, oldDocumentCategory.Prefix, oldPrefix).
				Delete(&models.CasbinRule{}).Error; err != nil {
				return fmt.Errorf("failed to delete obsolete Casbin rule: %w", err)
			}

			// Delete the obsolete role rule from the database
			if err := tx.Delete(&rule).Error; err != nil {
				return fmt.Errorf("failed to delete obsolete role rule: %w", err)
			}
		}

		// Step 2: Create new Casbin rules for the updated payload with new category and prefix
		var newDocumentCategory models.CategoryDocument
		if err := config.DB.Where("id = ?", payload.DocumentCategoryID).First(&newDocumentCategory).Error; err != nil {
			return fmt.Errorf("failed to find new category document: %w", err)
		}

		for _, rule := range payload.RoleHasRules {
			// Insert a new Casbin rule manually with the updated category and prefix
			newCasbinRule := models.CasbinRule{
				Ptype: "p",
				V0:    rule.RoleGuardName,
				V1:    "document",
				V2:    rule.Action,
				V3:    newDocumentCategory.Prefix,
				V4:    payload.Prefix,
			}
			if err := tx.Create(&newCasbinRule).Error; err != nil {
				return fmt.Errorf("failed to add new Casbin rule: %w", err)
			}
		}

		// Step 3: Update document_type record

		// Update the document type fields
		documentType.Name = payload.Name
		documentType.Prefix = payload.Prefix
		documentType.DocumentCategoryID = IntPtr(payload.DocumentCategoryID)
		documentType.UpdatedAt = time.Now()

		if err := tx.Save(&documentType).Error; err != nil {
			return fmt.Errorf("failed to update type document: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return &documentType, nil
}

// DeleteDocumentType soft deletes a document type by setting the deleted_at timestamp
func (s *DocumentTypeService) DeleteDocumentType(uuid string) error {
	currentTime := time.Now()

	// Check if document type with UUID exists
	var documentType models.DocumentType
	if err := config.DB.Where("uuid = ?", uuid).First(&documentType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document type not found")
		}
		return errors.New("failed to check document type")
	}

	// Soft delete by setting deleted_at timestamp
	if err := config.DB.Model(&documentType).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete document type")
	}

	return nil
}
