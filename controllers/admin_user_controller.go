package controllers

import (
	"backend-school/dto"
	"backend-school/helpers"
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Login handles user login and returns a JWT token along with the user's Casbin policies (v1 and v2 only)
func Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"data":       nil,
			"message":    "Cannot parse JSON",
		})
	}

	// Authenticate the user and get the token
	token, err := services.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"statusCode": fiber.StatusUnauthorized,
			"data":       nil,
			"message":    "Invalid credentials",
		})
	}

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Fetch policies directly related to the user
	userPolicies, err := enforcer.GetFilteredPolicy(0, req.Username)
	if err != nil {
		log.Printf("Error while fetching policies for user %s: %v", req.Username, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch user policies",
		})
	}

	// Fetch roles related to the user
	roles, err := enforcer.GetRolesForUser(req.Username)
	if err != nil {
		log.Printf("Error while fetching roles for user %s: %v", req.Username, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch user roles",
		})
	}

	// Fetch policies associated with each role
	var rolePolicies [][]string
	for _, role := range roles {
		rolePolicy, err := enforcer.GetFilteredPolicy(0, role)
		if err != nil {
			log.Printf("Error while fetching policies for role %s: %v", role, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"statusCode": fiber.StatusInternalServerError,
				"message":    "Failed to fetch role policies",
			})
		}
		rolePolicies = append(rolePolicies, rolePolicy...)
	}

	// Combine both user-specific policies and role-based policies
	allPolicies := append(userPolicies, rolePolicies...)

	// Extract v1 (resource) and v2 (action) from the policies
	var abilities []fiber.Map
	for _, policy := range allPolicies {
		if len(policy) >= 3 {
			abilities = append(abilities, fiber.Map{
				"resource": policy[1], // v1: resource
				"action":   policy[2], // v2: action
			})
		}
	}

	// Log the loaded policies and roles

	// Return the token along with the user's abilities (v1 and v2 only)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"data": fiber.Map{
			"token":   token,
			"ability": abilities, // Only v1 (resource) and v2 (action) in the response
		},
		"message": "Login successful",
	})
}

func GetUserData(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	// Use the fully qualified function name if it's in another package
	userData, err := services.GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "User not found",
			"data":       nil,
		})
	}

	return c.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User data retrieved successfully",
		"data":       userData,
	})
}

// GetUserDataAdmin retrieves user data if the user has admin access
// func GetUserDataAdmin(c *fiber.Ctx) error {
// 	// Get the username from the context (set by the JWT middleware)
// 	username := c.Locals("username").(string)

// 	// Get the Casbin enforcer
// 	enforcer := helpers.GetCasbinEnforcer()

// 	// Check if the user has access to the "/admin" resource using the "GET" action
// 	hasAccess, err := enforcer.Enforce(username, "/admin", "GET")
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"statusCode": fiber.StatusInternalServerError,
// 			"message":    "Failed to check access.",
// 		})
// 	}

// 	if !hasAccess {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"statusCode": fiber.StatusForbidden,
// 			"message":    "Forbidden: You don't have access to this resource",
// 		})
// 	}

// 	// Fetch user data after access is granted
// 	userData, err := services.GetUserByUsername(username)
// 	if err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"statusCode": fiber.StatusNotFound,
// 			"message":    "User not found",
// 			"data":       nil,
// 		})
// 	}

// 	// Return user data on success
// 	return c.JSON(fiber.Map{
// 		"statusCode": fiber.StatusOK,
// 		"message":    "User data retrieved successfully",
// 		"data":       userData,
// 	})
// }

func Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest

	// body := c.Body()

	// Parse and validate the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"data":       nil,
			"message":    "Invalid input",
		})
	}

	// Call the Register service to create the new user
	response, err := services.Register(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"data":       nil,
			"message":    err.Error(),
		})
	}

	// Assign the user a role in the Casbin rule table
	if err := services.AssignUserRole(req.Username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"data":       nil,
			"message":    "User created, but failed to assign role",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"data":       response,
		"message":    "User registered successfully",
	})
}

// ForgotPasswordHandler handles the request to initiate the password reset process
func ForgotPasswordHandler(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest

	// Parse the email from the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Cannot parse JSON",
		})
	}

	// Call the ForgotPassword service
	if err := services.ForgotPassword(req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Password reset email sent",
	})
}

// ResetPasswordHandler handles the request to reset the password
func ResetPasswordHandler(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest

	// Parse the reset token and new password from the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Cannot parse JSON",
		})
	}

	// Call the ResetPassword service
	if err := services.ResetPassword(req.Token, req.NewPassword); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Password updated successfully",
	})
}

// GetAllUsersPaginated handles fetching paginated users
func GetAllUsersPaginated(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "users" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "users", "read", "none", "none", "none")
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
	role := ctx.Query("role", "")
	email := ctx.Query("email", "")
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

	// Call the service to get paginated users
	users, paginationData, err := services.GetUsersPaginated(perPage, page, sortBy, sortDesc, role, email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch users",
		})
	}

	// Return the paginated data with a success response
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode":      fiber.StatusOK,
		"message":         "Users fetched successfully",
		"data":            users,
		"pagination_data": paginationData, // Contains current_page, total_pages, etc.
	})
}

// GetUserDetailByUUID handles the request to get user detail by UUID with Casbin rules (ptype g) included
func GetUserDetailByUUID(c *fiber.Ctx) error {
	// Get UUID from the route parameter
	userUUID := c.Params("uuid")

	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "read", "none", "none", "none")
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

	// Call the service to get user data by UUID using UserByAdmin model
	userData, err := services.GetUserByUUID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    err.Error(),
		})
	}

	// Retrieve the username associated with the user data
	targetUsername := userData.Username

	// Fetch Casbin policies related to the user (ptype "g", which indicates roles)
	userRoles, err := enforcer.GetFilteredGroupingPolicy(0, targetUsername) // Grouping policies (ptype "g")
	if err != nil {
		log.Printf("Error fetching Casbin policies for user %s: %v", targetUsername, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch user roles.",
		})
	}

	// Structure the roles with username and role
	var detailedRoles []fiber.Map
	for _, role := range userRoles {
		if len(role) >= 2 { // Ensure that we have both username and role
			detailedRoles = append(detailedRoles, fiber.Map{
				"username": role[0],
				"role":     role[1],
			})
		}
	}

	// Structure the response to include both user data and Casbin rules (ptype "g")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User data retrieved successfully",
		"data": fiber.Map{
			"user":  userData,      // User data
			"roles": detailedRoles, // Detailed Casbin roles (username and role)
		},
	})
}

// AddUserRoleByUUIDHandler handles the request to add a role to a user using UUID
func AddUserRoleByUUIDHandler(c *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "create", "none", "none", "none")
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
	// Get UUID from the route parameter
	userUUID := c.Params("uuid")

	// Get role from the request body
	var req struct {
		Role string `json:"role_guard_name"`
	}

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid input",
		})
	}

	// Check if role_guard_name is provided
	if req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "role_guard_name is required",
		})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid input",
		})
	}

	// Call the service to add the role to the user by UUID
	err = services.AddUserRoleByUUID(userUUID, req.Role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
		})
	}

	// Return success if the role was added
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Role added successfully",
	})
}

// DeleteUserRoleByUUIDHandler handles the request to delete a role from a user using UUID
func DeleteUserRoleByUUIDHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "delete", "none", "none", "none")
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
	// Get UUID from the route parameter
	userUUID := c.Params("uuid")

	// Get role_guard_name from the request body
	var req struct {
		RoleGuardName string `json:"role_guard_name"`
	}

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid input",
		})
	}

	// Check if role_guard_name is provided
	if req.RoleGuardName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "role_guard_name is required",
		})
	}

	// Call the service to delete the role from the user by UUID
	err = services.DeleteUserRoleByUUID(userUUID, req.RoleGuardName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
		})
	}

	// Return success if the role was deleted
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Role deleted successfully",
	})
}

// CreateUserByAdminHandler handles the request from an admin to create a user with a specific role
func CreateUserByAdminHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "create", "none", "none", "none")
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
	var reqAdmin dto.RegisterRequestAdmin

	// Parse the request body
	if err := c.BodyParser(&reqAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate the request body
	if err := helpers.ValidateStruct(&reqAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(), // This will return detailed validation error message
		})
	}

	// Buat objek RegisterRequest dari RegisterRequestAdmin
	req := dto.RegisterRequest{
		Fullname: reqAdmin.Fullname,
		Username: reqAdmin.Username,
		Password: reqAdmin.Password,
		Mobile:   reqAdmin.Mobile,
		Email:    reqAdmin.Email,
	}

	// Get the role_guard_name from the request body
	roleGuardName := reqAdmin.RoleGuardName
	if roleGuardName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Role guard name is required",
		})
	}

	// Call the service to create the user and assign the role
	response, err := services.CreateUserByAdmin(req, roleGuardName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "User created successfully",
		"data":       response,
	})
}

// UpdateUserByAdminHandler handles the request to update user details by admin
func UpdateUserByAdminHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "update", "none", "none", "none")
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
	var req dto.UpdateUserRequest

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Validate the request body
	if err := helpers.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
		})
	}

	// Get the UUID from the route parameter
	userUUID := c.Params("uuid")

	// Get role_guard_name from the request body
	var roleReq struct {
		RoleGuardName string `json:"role_guard_name"`
	}
	if err := c.BodyParser(&roleReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request for role_guard_name",
		})
	}

	// Call the service to update the user
	response, err := services.UpdateUserByAdmin(userUUID, req, roleReq.RoleGuardName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User updated successfully",
		"data":       response,
	})
}

// DeleteUserByAdminHandler handles the request to delete a user by admin
func DeleteUserByAdminHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "delete", "none", "none", "none")
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
	// Get the UUID from the route parameter
	userUUID := c.Params("uuid")

	// Call the service to delete the user
	err = services.DeleteUserByAdmin(userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User deleted successfully",
	})
}

// ActivateUserHandler handles the request to activate a user
func ActivateUserHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "update", "none", "none", "none")
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
	// Get the UUID from the route parameter
	userUUID := c.Params("uuid")

	// Call the service to activate the user
	err = services.ActivateUser(userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User activated successfully",
	})
}

// ActivateUserHandler handles the request to activate a user
func DeactivateUserHandler(c *fiber.Ctx) error {
	requesterUsername := c.Locals("username").(string)

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "users", "update", "none", "none", "none")
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
	// Get the UUID from the route parameter
	userUUID := c.Params("uuid")

	// Call the service to activate the user
	err = services.DeactivateUser(userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "User Deactivated successfully",
	})
}
