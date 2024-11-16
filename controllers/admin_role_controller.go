package controllers

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetAllRolesHandler handles the request to get all roles from the roles table
func GetAllRolesHandler(c *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "roles", "read", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}
	// Call the service to get all roles
	roles, err := services.GetAllRoles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve roles",
		})
	}

	// Return the list of roles in the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Roles retrieved successfully",
		"data":       roles,
	})
}

// GetPaginatedRolesHandler handles fetching paginated roles
func GetPaginatedRolesHandler(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "roles" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "roles", "read", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}

	// Default query parameters for pagination
	perPageStr := ctx.Query("perPage", "10")
	pageStr := ctx.Query("page", "1")
	sortBy := ctx.Query("sortBy", "id")           // Default sort by "id"
	sortDescStr := ctx.Query("sortDesc", "false") // Default sort in ascending order

	// Convert the query parameters to integers
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid perPage value",
		})
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid page value",
		})
	}

	sortDesc, err := strconv.ParseBool(sortDescStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid sortDesc value",
		})
	}

	// Call the service to get paginated roles
	roles, paginationData, err := services.GetPaginatedRoles(perPage, page, sortBy, sortDesc)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch roles",
		})
	}

	// Return the paginated data with a success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode":      fiber.StatusOK,
		"message":         "Roles fetched successfully",
		"data":            roles,
		"pagination_data": paginationData, // Contains current_page, total_pages, etc.
	})
}

// GetRoleByUUIDHandler handles fetching a role by its UUID and dynamically setting permissions based on actions in the role_has_rule table
func GetRoleByUUIDHandler(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "roles" resource using the "manage" action
	hasAccess, err := enforcer.Enforce(username, "all-content", "manage", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}

	// Get the UUID from the URL parameters
	uuid := ctx.Params("uuid")

	// Query the role by its UUID
	var role models.Role
	if err := config.DB.Where("uuid = ?", uuid).First(&role).Error; err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Role not found",
		})
	}

	// Define expected actions to ensure each rule_policy has all actions displayed
	expectedActions := []string{"create", "read", "update", "delete"}

	// Step 1: Retrieve all unique rule_policy and action combinations across all guard_names
	var allRules []models.RoleHasRule
	if err := config.DB.Find(&allRules).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve all role rules",
		})
	}

	// Step 2: Organize all rule policies with actions as a base structure
	uniquePolicies := make(map[string]map[string]bool)
	for _, rule := range allRules {
		if _, exists := uniquePolicies[rule.RulePolicy]; !exists {
			uniquePolicies[rule.RulePolicy] = make(map[string]bool)
		}
		uniquePolicies[rule.RulePolicy][rule.Action] = false // Initially set all actions to false
	}

	// Step 3: Retrieve only the rules associated with this role's guard_name
	var roleRules []models.RoleHasRule
	if err := config.DB.Where("role_guard_name = ?", role.GuardName).Find(&roleRules).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve role-specific rules",
		})
	}

	// Step 4: Update uniquePolicies with true for actions this role_guard_name has
	for _, rule := range roleRules {
		if actions, exists := uniquePolicies[rule.RulePolicy]; exists {
			actions[rule.Action] = true
		}
	}

	// Step 5: Convert uniquePolicies into the permissions structure for the response
	permissions := []fiber.Map{}
	for rulePolicy, actions := range uniquePolicies {
		actionMap := fiber.Map{}
		for _, action := range expectedActions {
			// Set each expected action, defaulting to false if not present
			actionMap[action] = actions[action]
		}
		permissions = append(permissions, fiber.Map{
			"rule_policy": rulePolicy,
			"action":      actionMap,
		})
	}

	// Prepare the final response structure
	response := fiber.Map{
		"data": fiber.Map{
			"name":            role.Name,
			"permissions":     permissions,
			"role_guard_name": role.GuardName,
		},
	}

	// Return the formatted response
	return ctx.Status(fiber.StatusOK).JSON(response)
}

// CreateRoleHandler handles the creation of a new role
func CreateRoleHandler(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "roles" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "roles", "create", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}
	// Define a struct to hold the request body
	type CreateRoleRequest struct {
		Name      string `json:"name" validate:"required"`
		GuardName string `json:"guard_name" validate:"required"`
	}

	// Parse the request body
	var request CreateRoleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate the required fields
	if request.Name == "" || request.GuardName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Role name and guard_name are required",
		})
	}

	// Call the service to create the role
	role, err := services.CreateRole(request.Name, request.GuardName)
	if err != nil {
		if err.Error() == "role with this name and guard_name already exists" {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"statusCode": fiber.StatusConflict,
				"message":    "Role with this name and guard_name already exists",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to create role",
		})
	}

	// Return the created role in the response
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Role created successfully",
		"data":       role,
	})
}

// UpdateRoleByUUIDHandler handles updating a role by its UUID
func UpdateRoleByUUIDHandler(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "roles", "update", "none", "none", "none")
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
	type UpdateRoleRequest struct {
		Name      string `json:"name" validate:"required"`
		GuardName string `json:"guard_name" validate:"required"`
	}

	// Parse the request body
	var request UpdateRoleRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate the required fields
	if request.Name == "" || request.GuardName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Role name and guard_name are required",
		})
	}

	// Call the service to update the role by UUID
	role, err := services.UpdateRoleByUUID(uuid, request.Name, request.GuardName)
	if err != nil {
		if err.Error() == "role not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Role not found",
			})
		}
		if err.Error() == "role with this name and guard_name already exists" {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"statusCode": fiber.StatusConflict,
				"message":    "Role with this name and guard_name already exists",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to update role",
		})
	}

	// Return the updated role in the response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Role updated successfully",
		"data":       role,
	})
}

// DeleteRoleByUUIDHandler handles the deletion of a role by its UUID
func DeleteRoleByUUIDHandler(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "roles", "delete", "none", "none", "none")
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

	// Call the service to delete the role by UUID
	err = services.DeleteRoleByUUID(uuid)
	if err != nil {
		if err.Error() == "invalid UUID format" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"statusCode": fiber.StatusBadRequest,
				"message":    "Invalid UUID format",
			})
		}
		if err.Error() == "role not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Role not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete role",
		})
	}

	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Role deleted successfully",
	})
}

// AddCasbinRuleHandler activates multiple Casbin rules based on the new payload structure
func AddCasbinRuleHandlerBulk(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "all-content", "manage", "none", "none", "none")
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
	var requestData struct {
		Data struct {
			Name          string                   `json:"name"`
			Permissions   []map[string]interface{} `json:"permissions"`
			RoleGuardName string                   `json:"role_guard_name"`
		} `json:"data"`
	}

	// Parse the JSON body into requestData
	if err := ctx.BodyParser(&requestData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Call the service to activate Casbin rules
	err = services.ActivateCasbinRulesBulk(requestData.Data.RoleGuardName, requestData.Data.Permissions)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to activate rules",
		})
	}

	// Return success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Rules activated successfully",
	})
}
