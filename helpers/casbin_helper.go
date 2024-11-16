package helpers

import (
	"log"
	"sync"
	"backend-school/config"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/gorm-adapter/v3"
  "backend-school/models" 
)

var enforcerInstance *casbin.Enforcer
var once sync.Once

// GetCasbinEnforcer creates or retrieves the Casbin enforcer instance.
func GetCasbinEnforcer() *casbin.Enforcer {
	once.Do(func() {
		// Initialize GORM adapter for Casbin, using GORM DB instance
		a, err := gormadapter.NewAdapterByDBWithCustomTable(config.DB, &models.CasbinRule{})
		if err != nil {
			log.Fatalf("Failed to initialize GORM adapter: %v", err)
		}

		// Load the Casbin model and policies from the database
		enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", a)
		if err != nil {
			log.Fatalf("Failed to load Casbin enforcer: %v", err)
		}

		// Load policies from DB
		err = enforcer.LoadPolicy()
		if err != nil {
			log.Fatalf("Failed to load Casbin policies: %v", err)
		}

		enforcerInstance = enforcer
	})

	return enforcerInstance
}
