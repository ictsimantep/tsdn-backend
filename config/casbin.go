package config

import (
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var Enforcer *casbin.Enforcer

func ConnectCasbin(DB *gorm.DB) {
	// Initialize the Casbin adapter with the database connection
	adapter, err := gormadapter.NewAdapterByDB(DB)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin adapter: %v", err)
	}

	// Use a single model file for Casbin configuration
	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to initialize Casbin enforcer with model: %v", err)
	}

	// Load policies from the database
	err = enforcer.LoadPolicy()
	if err != nil {
		log.Fatalf("Failed to load Casbin policies: %v", err)
	}

	// Set the global Enforcer variable
	Enforcer = enforcer
	log.Println("Casbin initialized successfully with a single model file.")
}
