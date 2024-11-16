package routes

import (
	"backend-school/controllers"
	"backend-school/middleware"
	"backend-school/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Group routes under /api
	api := app.Group("/api")
	// Initialize the TrafficIPService
	trafficIPService := services.NewTrafficIPService()
	trafficIPController := controllers.NewTrafficIPController(trafficIPService)
	api.Post("/traffic/static", trafficIPController.LogTrafficIP)

	// Auth routes
	api.Post("/auth/login", controllers.Login)
	api.Post("/auth/register", controllers.Register)
	api.Post("/auth/password/forgot", controllers.ForgotPasswordHandler)
	api.Post("/auth/password/reset", controllers.ResetPasswordHandler)
	api.Get("/auth/me", middleware.JWTMiddleware(), controllers.GetUserData)

	protectedUser := api.Group("/user", middleware.JWTMiddleware())
	documentControlController := controllers.NewDocumentControlController()
	protectedUser.Get("/document-control/list/internal", documentControlController.GetDocumentInternalControls) // List document controls with pagination
	protectedUser.Get("/document-control/list/external", documentControlController.GetDocumentExternalControls) // List document controls with pagination
	protectedUser.Post("/document-control", documentControlController.CreateDocumentControl)                    // Create a new document control
	protectedUser.Get("/document-control/:uuid", documentControlController.GetDocumentControlByUUID)            // Get a document control by UUID
	protectedUser.Put("/document-control/update/:uuid", documentControlController.UpdateDocumentControl)        // Update a document control by UUID
	protectedUser.Delete("/document-control/delete/:uuid", documentControlController.DeleteDocumentControl)

	// **Admin routes, protected by JWT Middleware, under /api/admin**
	protectedAdmin := api.Group("/admin", middleware.JWTMiddleware()) // Ensure middleware is applied here

	//ci = casbin implemented
	protectedAdmin.Get("/profiles", controllers.GetAllUsersPaginated)                     //ci
	protectedAdmin.Get("/profile/detail/:uuid", controllers.GetUserDetailByUUID)          //ci
	protectedAdmin.Post("/profile/create/:uuid", controllers.AddUserRoleByUUIDHandler)    //ci
	protectedAdmin.Post("/profile/delete/:uuid", controllers.DeleteUserRoleByUUIDHandler) //ci
	protectedAdmin.Get("/profile/roles", controllers.GetAllRolesHandler)

	protectedAdmin.Get("/roles", controllers.GetPaginatedRolesHandler)          //ci
	protectedAdmin.Get("/roles/detail/:uuid", controllers.GetRoleByUUIDHandler) //ci
	// protectedAdmin.Post("/roles/update/:uuid", controllers.UpdateRoleByUUIDHandler)   //ci
	protectedAdmin.Delete("/roles/delete/:uuid", controllers.DeleteRoleByUUIDHandler) //ci
	protectedAdmin.Post("/roles", controllers.CreateRoleHandler)                      //ci

	protectedAdmin.Post("/role-has-rule", controllers.CreateRoleHasRuleHandler)                      //ci
	protectedAdmin.Get("/role-has-rule", controllers.GetRoleHasRulesListHandler)                     //ci
	protectedAdmin.Get("/role-has-rule/paginated", controllers.GetPaginatedRoleHasRulesHandler)      //ci
	protectedAdmin.Put("/role-has-rule/update/:uuid", controllers.UpdateRoleHasRuleByUUIDHandler)    //ci
	protectedAdmin.Delete("/role-has-rule/delete/:uuid", controllers.DeleteRoleHasRuleByUUIDHandler) //ci
	protectedAdmin.Post("/rule/active", controllers.AddCasbinRuleHandler)                            //ci
	protectedAdmin.Post("/rule/deactive", controllers.DeleteCasbinRuleHandler)
	protectedAdmin.Post("/rule/active/bulk", controllers.AddCasbinRuleHandlerBulk) //ci
	protectedAdmin.Get("/rule-policy", controllers.GetUniqueRulePoliciesHandler)
	protectedAdmin.Get("/actions", controllers.GetActionsHandler)
	protectedAdmin.Post("/rule", controllers.CreateRoleHasRuleForAdminHandler)

	protectedAdmin.Get("/users", controllers.GetAllUsersPaginated)
	protectedAdmin.Get("/users/detail/:uuid", controllers.GetUserDetailByUUID)
	protectedAdmin.Post("/create/users", controllers.CreateUserByAdminHandler)
	protectedAdmin.Post("/update/users/:uuid", controllers.UpdateUserByAdminHandler)
	protectedAdmin.Delete("/delete/users/:uuid", controllers.DeleteUserByAdminHandler)
	protectedAdmin.Post("/activate/users/:uuid", controllers.ActivateUserHandler)
	protectedAdmin.Post("/deactivate/users/:uuid", controllers.DeactivateUserHandler)

	statusDocumentController := controllers.NewStatusDocumentController()
	protectedAdmin.Get("/status-document", statusDocumentController.GetStatusDocuments)                   // List status documents with pagination
	protectedAdmin.Post("/status-document", statusDocumentController.CreateStatusDocument)                // Create a new status document
	protectedAdmin.Get("/status-document/:uuid", statusDocumentController.GetStatusDocumentByUUID)        // Get a status document by UUID
	protectedAdmin.Put("/status-document/update/:uuid", statusDocumentController.UpdateStatusDocument)    // Update a status document by UUID
	protectedAdmin.Delete("/status-document/delete/:uuid", statusDocumentController.DeleteStatusDocument) // Delete a status document by UUID

	categoryDocumentController := controllers.NewCategoryDocumentController()
	protectedAdmin.Get("/category-document", categoryDocumentController.GetCategoryDocuments)                // List category documents with pagination
	protectedAdmin.Post("/category-document", categoryDocumentController.CreateCategoryDocument)             // Create a new category document
	protectedAdmin.Get("/category-document/:uuid", categoryDocumentController.GetCategoryDocumentByUUID)     // Get a category document by UUID
	protectedAdmin.Put("/category-document/update/:uuid", categoryDocumentController.UpdateCategoryDocument) // Update a category document by UUID
	protectedAdmin.Delete("/category-document/delete/:uuid", categoryDocumentController.DeleteCategoryDocument)

	documentTypeController := controllers.NewDocumentTypeController()
	protectedAdmin.Get("/document-type", documentTypeController.GetDocumentTypes)                   // List document types with pagination
	protectedAdmin.Post("/document-type", documentTypeController.CreateDocumentType)                // Create a new document type
	protectedAdmin.Get("/document-type/:uuid", documentTypeController.GetDocumentTypeByUUID)        // Get a document type by UUID
	protectedAdmin.Put("/document-type/update/:uuid", documentTypeController.UpdateDocumentType)    // Update a document type by UUID
	protectedAdmin.Delete("/document-type/delete/:uuid", documentTypeController.DeleteDocumentType) // Delete a document type by UUID

	protectedAdmin.Get("/role-action-master", categoryDocumentController.GetRolesAndActions)
	// Delete a document control by UUID

}
