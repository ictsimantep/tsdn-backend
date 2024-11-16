package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleWithCasbinRules holds a role and its associated Casbin rules
type RoleWithCasbinRules struct {
	Role  models.Role    `json:"role"`
	Rules []RuleResponse `json:"rules"` // Change this line to use RuleResponse
}

type RuleResponse struct {
	Role       string `json:"role"`
	RulePolicy string `json:"rule_policy"`
	Action     string `json:"action"`
}

// PaginationData holds pagination information
type PaginationData struct {
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
}

// GetAllRoles retrieves all roles from the roles table.
func GetAllRoles() ([]models.Role, error) {
	var roles []models.Role

	// Query the database to get all roles
	if err := config.DB.Find(&roles).Error; err != nil {
		return nil, err
	}

	// Return the list of roles
	return roles, nil
}

// GetPaginatedRoles retrieves a paginated list of roles from the database
func GetPaginatedRoles(perPage, page int, sortBy string, sortDesc bool) ([]models.Role, PaginationData, error) {
	var roles []models.Role
	var totalRecords int64

	// Calculate the offset for pagination
	offset := (page - 1) * perPage

	// Count total records in the roles table
	if err := config.DB.Model(&models.Role{}).Count(&totalRecords).Error; err != nil {
		return nil, PaginationData{}, err
	}

	// Set sorting order (ASC/DESC)
	order := sortBy
	if sortDesc {
		order += " DESC"
	} else {
		order += " ASC"
	}

	// Fetch paginated roles with sorting
	if err := config.DB.Order(order).Limit(perPage).Offset(offset).Find(&roles).Error; err != nil {
		return nil, PaginationData{}, err
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Prepare pagination data
	paginationData := PaginationData{
		TotalRecords: int(totalRecords),
		TotalPages:   totalPages,
		CurrentPage:  page,
		PageSize:     perPage,
	}

	return roles, paginationData, nil
}

// GetRoleWithCasbinRulesByUUID retrieves a role by its UUID and its associated Casbin rules
func GetRoleWithCasbinRulesByUUID(uuidStr string) (*RoleWithCasbinRules, error) {
	var role models.Role

	// Parse the UUID string to a UUID type
	uuidVal, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	// Query the database for the role by UUID
	if err := config.DB.Where("uuid = ?", uuidVal).First(&role).Error; err != nil {
		return nil, err
	}

	// Fetch the Casbin rules associated with this role using GuardName
	var rules []models.CasbinRule
	if err := config.DB.Where("v0 = ?", role.GuardName).Where("ptype = ?", "p").Find(&rules).Error; err != nil {
		return nil, err
	}

	// Transform the rules into the new format
	var transformedRules []RuleResponse
	for _, rule := range rules {
		transformedRules = append(transformedRules, RuleResponse{
			Role:       rule.V0,
			RulePolicy: rule.V1,
			Action:     rule.V2,
		})
	}

	return &RoleWithCasbinRules{Role: role, Rules: transformedRules}, nil
}

// CreateRole creates a new role in the database
func CreateRole(name string, guardName string) (*models.Role, error) {
	// Check if a role with the same name and guard_name already exists
	var existingRole models.Role
	if err := config.DB.Where("name = ? AND guard_name = ?", name, guardName).First(&existingRole).Error; err == nil {
		log.Printf("Role with this name and guard_name already exists: %s, %s", name, guardName)
		return nil, errors.New("role with this name and guard_name already exists")
	}

	// Get the database connection and handle the error
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "roles", "id", "role_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return nil, fmt.Errorf("failed to reset sequence: %w", err)
	}

	// Create a new role
	role := models.Role{
		Name:      name,
		GuardName: guardName,
	}

	// Insert the role into the database
	if err := config.DB.Create(&role).Error; err != nil {
		log.Printf("Failed to create role in database: %v", err)
		return nil, err
	}

	log.Printf("Role created successfully: %+v", role)
	return &role, nil
}

// UpdateRoleByUUID updates the role's name and guard_name based on its UUID
func UpdateRoleByUUID(uuidStr string, name string, guardName string) (*models.Role, error) {
	var role models.Role

	// Parse the UUID string to a UUID type
	uuidVal, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	// Find the role by UUID
	if err := config.DB.Where("uuid = ?", uuidVal).First(&role).Error; err != nil {
		return nil, errors.New("role not found")
	}

	// Check if a different role already exists with the same name and guard_name
	var existingRole models.Role
	if err := config.DB.Where("name = ? AND guard_name = ? AND uuid != ?", name, guardName, uuidVal).First(&existingRole).Error; err == nil {
		return nil, errors.New("role with this name and guard_name already exists")
	}

	// Update the role fields
	role.Name = name
	role.GuardName = guardName

	// Save the updated role to the database
	if err := config.DB.Save(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRoleByUUID deletes a role by its UUID from the database
func DeleteRoleByUUID(uuidStr string) error {
	// Parse the UUID string to a UUID type
	uuidVal, err := uuid.Parse(uuidStr)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Find the role by UUID
	var role models.Role
	if err := config.DB.Where("uuid = ?", uuidVal).First(&role).Error; err != nil {
		return errors.New("role not found")
	}

	// Delete the role
	if err := config.DB.Delete(&role).Error; err != nil {
		return errors.New("failed to delete role")
	}

	return nil
}

// ActivateCasbinRulesBulk activates multiple Casbin rules for a given role based on the provided permissions payload structure
func ActivateCasbinRulesBulk(roleGuardName string, permissions []map[string]interface{}) error {
	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Iterate over the permissions array
	for _, permissionData := range permissions {
		rulePolicy, _ := permissionData["rule_policy"].(string)
		actions, _ := permissionData["action"].(map[string]interface{})

		// Iterate over each action in the action map to handle it dynamically
		for action, allowed := range actions {
			// Check if the action-policy combination already exists in role_has_rule
			var existingRule models.RoleHasRule
			err := config.DB.Where("role_guard_name = ? AND rule_policy = ? AND action = ?", roleGuardName, rulePolicy, action).First(&existingRule).Error

			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				// If the rule does not exist, create it in the role_has_rule table
				newRule := models.RoleHasRule{
					RoleGuardName: roleGuardName,
					RulePolicy:    rulePolicy,
					Action:        action,
				}

				// Insert the new rule
				if err := config.DB.Create(&newRule).Error; err != nil {
					log.Printf("Failed to create rule in role_has_rule for %s on %s with action %s: %v", roleGuardName, rulePolicy, action, err)
					return err
				}
			} else if err != nil {
				log.Printf("Error checking existing rule in role_has_rule: %v", err)
				return err
			}

			// Add or remove policy in Casbin based on the allowed status
			if allowed.(bool) {
				// Add policy if action is allowed
				_, err := enforcer.AddPolicy(roleGuardName, rulePolicy, action)
				if err != nil {
					log.Printf("Failed to add '%s' rule for %s on %s: %v", action, roleGuardName, rulePolicy, err)
					return err
				}
			} else {
				// Remove policy if action is not allowed
				_, err := enforcer.RemovePolicy(roleGuardName, rulePolicy, action)
				if err != nil {
					log.Printf("Failed to remove '%s' rule for %s on %s: %v", action, roleGuardName, rulePolicy, err)
					return err
				}
			}
		}
	}

	// Save policies to the backend (e.g., database)
	err := enforcer.SavePolicy()
	if err != nil {
		log.Printf("Failed to save Casbin policy: %v", err)
		return err
	}

	return nil
}
