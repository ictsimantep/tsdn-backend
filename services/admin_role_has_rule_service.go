package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"errors"
	"log"
	"math"

	"github.com/google/uuid"
)

// RoleWithRules represents the structure for roles with associated rules
type RoleWithRules struct {
	RoleGuardName string        `json:"role_guard_name"`
	Rules         []RuleSummary `json:"rules"`
}

// RuleSummary represents the rule details for each role
type RuleSummary struct {
	RulePolicy string `json:"rule_policy"`
	Action     string `json:"action"`
	Active     bool   `json:"active"`
}

// RoleHasRuleSummary represents the structure for each role with its associated rule
type RoleHasRuleSummary struct {
	ID            uint   `json:"id"`
	UUID          string `json:"uuid"`
	RoleGuardName string `json:"role_guard_name"`
	RulePolicy    string `json:"rule_policy"`
	Action        string `json:"action"`
}

// GetPaginatedRoleHasRules retrieves a paginated list of role_has_rules with ID and UUID
func GetPaginatedRoleHasRules(page int, pageSize int) ([]RoleHasRuleSummary, PaginationData, error) {
	var roleRules []models.RoleHasRule
	var roleHasRuleSummaries []RoleHasRuleSummary
	var totalRecords int64

	// Calculate the offset for pagination
	offset := (page - 1) * pageSize

	// Count total records in the role_has_rules table
	if err := config.DB.Model(&models.RoleHasRule{}).Count(&totalRecords).Error; err != nil {
		return nil, PaginationData{}, err
	}

	// Fetch paginated records
	if err := config.DB.Limit(pageSize).Offset(offset).Find(&roleRules).Error; err != nil {
		return nil, PaginationData{}, err
	}

	// Iterate over the fetched records and convert to summary format including ID and UUID
	for _, rule := range roleRules {
		roleHasRuleSummaries = append(roleHasRuleSummaries, RoleHasRuleSummary{
			ID:            rule.ID,
			UUID:          rule.UUID.String(),
			RoleGuardName: rule.RoleGuardName,
			RulePolicy:    rule.RulePolicy,
			Action:        rule.Action,
		})
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	// Prepare pagination data
	paginationData := PaginationData{
		TotalRecords: int(totalRecords),
		TotalPages:   totalPages,
		CurrentPage:  page,
		PageSize:     pageSize,
	}

	return roleHasRuleSummaries, paginationData, nil
}

// CreateRoleWithRules processes the payload, adds entries to roles, role_has_rules, and Casbin policies.
func CreateRoleHasRules(roleGuardName string, roleName string, rules []map[string]interface{}) ([]models.RoleHasRule, error) {
	var createdRules []models.RoleHasRule

	// Start a database transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Step 1: Create the Role in `roles` table (if it doesn’t exist)
	var role models.Role
	if err := tx.Where("guard_name = ?", roleGuardName).First(&role).Error; err != nil {
		// If role doesn't exist, create it
		role = models.Role{
			Name:      roleName,
			GuardName: roleGuardName,
		}
		if err := tx.Create(&role).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to create role in roles table")
		}
	}

	// Step 2: Process each rule in the `rules` array and add entries to `role_has_rules`
	for _, ruleData := range rules {
		// Extract rule_policy
		rulePolicy, ok := ruleData["rule_policy"].(string)
		if !ok || rulePolicy == "" {
			tx.Rollback()
			return nil, errors.New("invalid or missing rule_policy")
		}

		// Extract actions map
		actions, ok := ruleData["action"].(map[string]interface{})
		if !ok {
			tx.Rollback()
			return nil, errors.New("invalid or missing actions for rule")
		}

		// Process each action (create, read, update, delete)
		for action, enabled := range actions {
			isEnabled, ok := enabled.(bool)
			if !ok {
				tx.Rollback()
				return nil, errors.New("action value must be boolean")
			}

			// Only create the rule if the action is enabled (set to true)
			if isEnabled {
				// Define the role_has_rule entry
				rule := models.RoleHasRule{
					RoleGuardName: roleGuardName,
					RulePolicy:    rulePolicy,
					Action:        action,
				}

				// Check if the rule already exists in `role_has_rules`
				var existingRule models.RoleHasRule
				if err := tx.Where("role_guard_name = ? AND rule_policy = ? AND action = ?", roleGuardName, rulePolicy, action).
					First(&existingRule).Error; err == nil {
					log.Printf("Rule with this role_guard_name, rule_policy, and action already exists: %s, %s, %s", roleGuardName, rulePolicy, action)
					continue // Skip adding duplicate rule
				}

				// Insert the rule into the `role_has_rules` table
				if err := tx.Create(&rule).Error; err != nil {
					tx.Rollback()
					return nil, errors.New("failed to create rule in role_has_rules table")
				}
				createdRules = append(createdRules, rule)

				// Step 3: Add rule to Casbin as a policy
				success, err := enforcer.AddPolicy(roleGuardName, rulePolicy, action)
				if err != nil || !success {
					tx.Rollback()
					return nil, errors.New("failed to add policy to Casbin")
				}
			}
		}
	}

	// Commit the transaction for `roles` and `role_has_rules`
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Save policies to the Casbin backend (e.g., database)
	if err := enforcer.SavePolicy(); err != nil {
		return nil, errors.New("failed to save Casbin policies to storage")
	}

	return createdRules, nil
}

// GetRoleHasRulesList retrieves the list of roles with associated rules
func GetRoleHasRulesList() ([]RoleWithRules, error) {
	var roleRules []models.RoleHasRule
	var roleWithRulesList []RoleWithRules

	// Query the role_has_rule table
	if err := config.DB.Find(&roleRules).Error; err != nil {
		return nil, err
	}

	// Map to group rules by role_guard_name
	roleMap := make(map[string][]RuleSummary)

	// Iterate over the fetched rules and group them by role_guard_name
	for _, rule := range roleRules {
		ruleSummary := RuleSummary{
			RulePolicy: rule.RulePolicy,
			Action:     rule.Action,
		}
		roleMap[rule.RoleGuardName] = append(roleMap[rule.RoleGuardName], ruleSummary)
	}

	// Convert the map into the final output structure
	for roleGuardName, rules := range roleMap {
		roleWithRulesList = append(roleWithRulesList, RoleWithRules{
			RoleGuardName: roleGuardName,
			Rules:         rules,
		})
	}

	return roleWithRulesList, nil
}

// UpdateRoleHasRuleByUUID updates a role_has_rule record by its UUID
func UpdateRoleHasRuleByUUID(uuidStr string, roleGuardName string, rulePolicy string, action string) (*models.RoleHasRule, error) {
	var roleHasRule models.RoleHasRule

	// Parse the UUID string to a UUID type
	uuidVal, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	// Find the record by UUID
	if err := config.DB.Where("uuid = ?", uuidVal).First(&roleHasRule).Error; err != nil {
		return nil, errors.New("role_has_rule not found")
	}

	// Update the fields
	roleHasRule.RoleGuardName = roleGuardName
	roleHasRule.RulePolicy = rulePolicy
	roleHasRule.Action = action

	// Save the updated record
	if err := config.DB.Save(&roleHasRule).Error; err != nil {
		return nil, err
	}

	return &roleHasRule, nil
}
func DeleteRoleHasRuleByUUID(uuidStr string) error {
	var roleHasRule models.RoleHasRule

	// Parse the UUID string to a UUID type
	uuidVal, err := uuid.Parse(uuidStr)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Find the record by UUID
	if err := config.DB.Where("uuid = ?", uuidVal).First(&roleHasRule).Error; err != nil {
		return errors.New("role_has_rule not found")
	}

	// Delete the record
	if err := config.DB.Delete(&roleHasRule).Error; err != nil {
		return err
	}

	return nil
}

// CheckRuleInCasbin checks if a rule exists in the casbin_rule table
func CheckRuleInCasbin(roleGuardName string, rulePolicy string, action string) (bool, error) {
	var count int64

	// Query to check if the rule exists in the casbin_rule table
	if err := config.DB.Model(&models.CasbinRule{}).
		Where("v0 = ? AND v1 = ? AND v2 = ?", roleGuardName, rulePolicy, action).
		Count(&count).Error; err != nil {
		return false, err
	}

	// If count is greater than 0, the rule exists
	return count > 0, nil
}

// AddCasbinRule adds a new rule to the Casbin enforcer and stores it in the database
func AddCasbinRule(roleGuardName string, rulePolicy string, action string, category string, typeCR string) (bool, error) {
	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Add a new policy (roleGuardName, rulePolicy, action) to Casbin
	success, err := enforcer.AddPolicy(roleGuardName, rulePolicy, action, category, typeCR, "none")
	if err != nil {
		return false, err
	}

	// If the policy was successfully added, return true
	return success, nil
}

// DeleteCasbinRule removes a rule from the Casbin enforcer and the database
func DeleteCasbinRule(roleGuardName string, rulePolicy string, action string, category string, typeCR string) (bool, error) {
	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Remove the policy (roleGuardName, rulePolicy, action) from Casbin
	success, err := enforcer.RemovePolicy(roleGuardName, rulePolicy, action, category, typeCR)
	if err != nil {
		return false, err
	}

	// If the policy was successfully removed, return true
	return success, nil
}

// GetUniqueRulePolicies retrieves a unique list of rule policies from the role_has_rules table
func GetUniqueRulePolicies() ([]string, error) {
	var rulePolicies []string

	// Query to get distinct rule policies from the role_has_rule table
	if err := config.DB.Model(&models.RoleHasRule{}).Distinct("rule_policy").Pluck("rule_policy", &rulePolicies).Error; err != nil {
		return nil, err
	}

	return rulePolicies, nil
}

// GetUniqueActions retrieves a unique list of actions from the role_has_rules table
func GetUniqueActions() ([]string, error) {
	var actions []string

	// Query to get distinct actions from the role_has_rules table
	if err := config.DB.Model(&models.RoleHasRule{}).Distinct("action").Pluck("action", &actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

func CreateRoleHasRuleForAdmin(rulePolicy string, action string) (*models.RoleHasRule, error) {
	roleGuardName := "admin" // Fixed role_guard_name

	// Start a database transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Check if the rule already exists in the role_has_rule table for "admin"
	var existingRule models.RoleHasRule
	if err := tx.Where("role_guard_name = ? AND rule_policy = ? AND action = ?", roleGuardName, rulePolicy, action).
		First(&existingRule).Error; err == nil {
		log.Printf("Rule with this role_guard_name, rule_policy, and action already exists: %s, %s, %s", roleGuardName, rulePolicy, action)
		tx.Rollback() // Rollback the transaction
		return nil, errors.New("rule with this role_guard_name, rule_policy, and action already exists")
	}

	// Define the new rule entry
	rule := models.RoleHasRule{
		RoleGuardName: roleGuardName,
		RulePolicy:    rulePolicy,
		Action:        action,
	}

	// Insert the new rule into the role_has_rule table
	if err := tx.Create(&rule).Error; err != nil {
		tx.Rollback() // Rollback on failure
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &rule, nil
}
