package config

import "os"

// SecretKey is used for signing JWT tokens
var SecretKey = os.Getenv("SECRET_KEY")
