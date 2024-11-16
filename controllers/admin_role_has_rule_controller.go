package controllers

import (
	"backend-school/helpers"
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetActionsHandler retrieves a unique list of actions
func GetActionsHandler(ctx *fiber.Ctx) error {
	// Fetch the actions from the service layer
	actions, err := services.GetUniqueActions()
	if err != nil {
		log.Printf("Error retrieving actions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve actions",
		})
	}

	// Return the actions in the response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Actions fetched successfully",
		"data":       actions,
	})
}

// GetRulePoliciesHandler retrieves a unique list of rule policies
func GetRulePoliciesHandler(ctx *fiber.Ctx) error {
	// Fetch the rule policies from the service layer
	rulePolicies, err := services.GetUniqueRulePolicies()
	if err != nil {
		log.Printf("Error retrieving rule policies: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve rule policies",
		})
	}

	// Return the rule policies in the response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Rule policies fetched successfully",
		"data":       rulePolicies,
	})
}

// CreateRoleHasRuleHandler handles the creation of a new role with multiple rules and actions
func CreateRoleHasRuleHandler(ctx *fiber.Ctx) error {
	// Define a struct to hold the request body with the new format
	type CreateRoleHasRuleRequest struct {
		RoleGuardName string                   `json:"role_guard_name" validate:"required"`
		Rules         []map[string]interface{} `json:"rules" validate:"required"`
	}

	// Parse the request body
	var request CreateRoleHasRuleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate required fields
	if request.RoleGuardName == "" || request.RoleGuardName == "" || len(request.Rules) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "RoleGuardName, RoleName, and Rules are required",
		})
	}

	// Call the service to create the role, rules, and Casbin policies
	createdRules, err := services.CreateRoleHasRules(request.RoleGuardName, request.RoleGuardName, request.Rules)
	if err != nil {
		log.Printf("Error creating role and rules: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to create role and rules",
		})
	}

	// Return the created rules in the response
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Role and rules created successfully",
		"data":       createdRules,
	})
}

// GetRoleHasRulesListHandler handles fetching the list of roles with associated rules
func GetRoleHasRulesListHandler(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "read", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}

	// Call the service to get the roles with rules
	roleWithRulesList, err := services.GetRoleHasRulesList()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve roles with rules",
		})
	}

	// Iterate over each role's rules and check against Casbin's rules (assuming it's in a database table)
	for i, role := range roleWithRulesList {
		for j, rule := range role.Rules {
			// Check if the rule exists in casbin_rule (you need to implement the function CheckRuleInCasbin)
			existsInCasbin, err := services.CheckRuleInCasbin(role.RoleGuardName, rule.RulePolicy, rule.Action)
			if err != nil {
				log.Printf("Error checking rule in Casbin: %v", err)
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"statusCode": fiber.StatusInternalServerError,
					"message":    "Failed to check rule in Casbin.",
				})
			}

			// Add the active field based on the check result
			roleWithRulesList[i].Rules[j].Active = existsInCasbin
		}
	}

	// Return the result
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Roles with rules fetched successfully",
		"data":       roleWithRulesList,
	})
}

// GetPaginatedRoleHasRulesHandler handles fetching the paginated list of role_has_rules
func GetPaginatedRoleHasRulesHandler(ctx *fiber.Ctx) error {
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "read", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Get query parameters for pagination
	pageStr := ctx.Query("page", "1")          // Default to page 1
	pageSizeStr := ctx.Query("pageSize", "10") // Default page size 10

	// Convert the query parameters to integers
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid page value",
		})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid pageSize value",
		})
	}

	// Call the service to get the paginated roles with rules
	roleHasRuleSummaries, paginationData, err := services.GetPaginatedRoleHasRules(page, pageSize)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve paginated roles with rules",
		})
	}

	// Return the paginated result along with pagination metadata
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode":      fiber.StatusOK,
		"message":         "Paginated roles with rules fetched successfully",
		"data":            roleHasRuleSummaries,
		"pagination_data": paginationData,
	})
}

// UpdateRoleHasRuleByUUIDHandler handles updating a role_has_rule by its UUID
func UpdateRoleHasRuleByUUIDHandler(ctx *fiber.Ctx) error {
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "update", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Get the UUID from the URL parameters
	uuid := ctx.Params("uuid")

	// Define a struct to hold the request body
	type UpdateRoleHasRuleRequest struct {
		RoleGuardName string `json:"role_guard_name" validate:"required"`
		RulePolicy    string `json:"rule_policy" validate:"required"`
		Action        string `json:"action" validate:"required"`
	}

	// Parse the request body
	var request UpdateRoleHasRuleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate required fields
	if request.RoleGuardName == "" || request.RulePolicy == "" || request.Action == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "RoleGuardName, RulePolicy, and Action are required",
		})
	}

	// Call the service to update the record by UUID
	updatedRoleHasRule, err := services.UpdateRoleHasRuleByUUID(uuid, request.RoleGuardName, request.RulePolicy, request.Action)
	if err != nil {
		if err.Error() == "invalid UUID format" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    "Invalid UUID format",
			})
		}
		if err.Error() == "role_has_rule not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "RoleHasRule not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to update RoleHasRule",
		})
	}

	// Return the updated record in the response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "RoleHasRule updated successfully",
		"data":       updatedRoleHasRule,
	})
}

// DeleteRoleHasRuleByUUIDHandler handles deleting a role_has_rule by its UUID
func DeleteRoleHasRuleByUUIDHandler(ctx *fiber.Ctx) error {
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "delete", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Get the UUID from the URL parameters
	uuid := ctx.Params("uuid")

	// Call the service to delete the rule by UUID
	err = services.DeleteRoleHasRuleByUUID(uuid)
	if err != nil {
		if err.Error() == "invalid UUID format" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    "Invalid UUID format",
			})
		}
		if err.Error() == "role_has_rule not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "RoleHasRule not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete RoleHasRule",
		})
	}

	// Return a success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "RoleHasRule deleted successfully",
	})
}

// AddCasbinRuleHandler handles adding a new rule to Casbin
func AddCasbinRuleHandler(ctx *fiber.Ctx) error {
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "create", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Define a struct to hold the request body
	type AddCasbinRuleRequest struct {
		RoleGuardName string `json:"role_guard_name" validate:"required"`
		RulePolicy    string `json:"rule_policy" validate:"required"`
		Action        string `json:"action" validate:"required"`
		Category      string `json:"category"`
		TypeCR        string `json:"type"`
	}

	// Parse the request body
	var request AddCasbinRuleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate required fields
	if request.RoleGuardName == "" || request.RulePolicy == "" || request.Action == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "RoleGuardName, RulePolicy, and Action are required",
		})
	}

	// Call the service to add the rule to Casbin
	success, err := services.AddCasbinRule(request.RoleGuardName, request.RulePolicy, request.Action, request.Category, request.TypeCR)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to add rule to Casbin",
		})
	}

	if !success {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"statusCode": fiber.StatusConflict,
			"message":    "Rule already exists in Casbin",
		})
	}

	// Return success response
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Rule added successfully",
	})
}

// DeleteCasbinRuleHandler handles deleting a rule from Casbin
func DeleteCasbinRuleHandler(ctx *fiber.Ctx) error {
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "delete", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Define a struct to hold the request body
	type DeleteCasbinRuleRequest struct {
		RoleGuardName string `json:"role_guard_name" validate:"required"`
		RulePolicy    string `json:"rule_policy" validate:"required"`
		Action        string `json:"action" validate:"required"`
		Category      string `json:"category"`
		TypeCR        string `json:"type"`
	}

	// Parse the request body
	var request DeleteCasbinRuleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate required fields
	if request.RoleGuardName == "" || request.RulePolicy == "" || request.Action == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "RoleGuardName, RulePolicy, and Action are required",
		})
	}

	// Call the service to delete the rule from Casbin
	success, err := services.DeleteCasbinRule(request.RoleGuardName, request.RulePolicy, request.Action, request.Category, request.TypeCR)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete rule from Casbin",
		})
	}

	if !success {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Rule not found in Casbin",
		})
	}

	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Rule deleted successfully",
	})
}

// GetUniqueRulePoliciesHandler handles fetching a unique list of rule policies
func GetUniqueRulePoliciesHandler(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view rule policies
	hasAccess, err := enforcer.Enforce(requesterUsername, "rules", "read", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}

	// Call the service to retrieve the unique rule policies
	rulePolicies, err := services.GetUniqueRulePolicies()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve rule policies",
		})
	}

	// Return the unique rule policies
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Rule policies fetched successfully",
		"data":       rulePolicies,
	})
}

// CreateRoleHasRuleForAdminHandler handles the creation of a new rule for admin in role_has_rule
func CreateRoleHasRuleForAdminHandler(ctx *fiber.Ctx) error {
	// Define a struct to hold the request body
	type CreateRuleRequest struct {
		RulePolicy string `json:"rule_policy" validate:"required"`
		Action     string `json:"action" validate:"required"`
	}

	// Parse the request body
	var request CreateRuleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate required fields
	if request.RulePolicy == "" || request.Action == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "RulePolicy and Action are required",
		})
	}

	// Call the service to create the rule for admin
	rule, err := services.CreateRoleHasRuleForAdmin(request.RulePolicy, request.Action)
	if err != nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"statusCode": fiber.StatusConflict,
			"message":    err.Error(),
		})
	}

	// Return the created rule in the response
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Rule created successfully",
		"data":       rule,
	})
}
