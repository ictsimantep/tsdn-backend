package services

import (
	"backend-school/config"
	"backend-school/dto"
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/templates"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-gomail/gomail"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Load JWT_SECRET from environment variable (SECRET_KEY)
var jwtSecret = []byte(os.Getenv("SECRET_KEY"))

// Login handles user login by verifying credentials and returning a JWT token.
func Login(username, password string) (string, error) {
	var user models.User

	// Retrieve the user by username from the database
	if err := config.DB.Where("username = ?", username).Where("deleted_at", nil).First(&user).Error; err != nil {
		return "", errors.New("invalid username or password")
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	// Generate a new JWT token
	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(), // Token expiration
		"iat":      time.Now().Unix(),                     // Issued at
		"nbf":      time.Now().Unix(),                     // Not before
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func HasAccess(username, path, method string) (bool, error) {
	// Check access using Casbin
	allowed, err := config.Enforcer.Enforce(username, path, method)
	if err != nil {
		return false, err
	}

	return allowed, nil
}

// Register handles user registration.
func Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	var existingUser models.UserRegister

	// Check if the user already exists by username
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create the new user model
	user := models.UserRegister{
		Fullname: req.Fullname,
		Username: req.Username,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Save the user to the database
	if err := config.DB.Create(&user).Error; err != nil {
		return nil, errors.New("could not create user, please try again")
	}

	// Create the response
	response := &dto.RegisterResponse{
		ID:       user.ID,
		UUID:     user.UUID,
		Fullname: user.Fullname,
		Username: user.Username,
		Mobile:   user.Mobile,
		Email:    user.Email,
		Message:  "User registered successfully",
	}

	return response, nil
}

func AssignUserRole(username string) error {
	rule := models.CasbinRule{
		Ptype: "g",
		V0:    username,
		V1:    "user",
	}

	if err := config.DB.Create(&rule).Error; err != nil {
		return errors.New("failed to assign user role")
	}

	return nil
}

// GetUserByUsername retrieves user data by username and returns it as a UserGetData object
func GetUserByUsername(username string) (*models.UserGetData, error) {
	var user models.UserGetData

	// Query the database for the user by username
	if err := config.DB.Where("username = ?", username).Where("deleted_at", nil).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Return the user details
	return &user, nil
}

// ForgotPassword handles the process of generating a password reset token and sending it to the user's email.
func ForgotPassword(email string) error {
	var user models.UserGetData

	// Check if the user exists in the database by email
	if err := config.DB.Where("email = ?", email).Where("deleted_at", nil).First(&user).Error; err != nil {
		return errors.New("email not found")
	}

	// Generate a secure reset token
	resetToken, err := generateResetToken()
	if err != nil {
		return errors.New("failed to generate reset token")
	}

	// Create a new entry in the PasswordResetToken table
	tokenExpiry := time.Now().Add(1 * time.Hour) // Token valid for 1 hour
	resetEntry := models.PasswordResetToken{
		Email:     user.Email,
		Token:     resetToken,
		ExpiresAt: tokenExpiry,
	}

	// Save the token in the PasswordResetToken table
	if err := config.DB.Create(&resetEntry).Error; err != nil {
		return errors.New("failed to save reset token")
	}

	// Load the reset URL base from the environment variable
	resetURLBase := os.Getenv("RESET_PASSWORD_URL")
	if resetURLBase == "" {
		return errors.New("reset URL not configured in environment variables")
	}

	// Create the full reset URL
	resetURL := fmt.Sprintf("%s?token=%s", resetURLBase, resetToken)

	// Send the reset email using Mailtrap
	err = sendResetEmail(user.Email, resetURL)
	if err != nil {
		return errors.New("failed to send reset email")
	}

	return nil
}

func sendResetEmail(email, resetLink string) error {
	// Replace {{RESET_LINK}} with the actual reset link in the template
	bodyWithLink := strings.Replace(templates.ResetEmailTemplate, "{{RESET_LINK}}", resetLink, -1)

	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@school.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", bodyWithLink)

	// Configure the Mailtrap SMTP settings (ensure credentials are correct)
	d := gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, "0e9340db84cbec", "f325d2b3ca4dcc")

	// Attempt to send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// generateResetToken generates a random reset token
func generateResetToken() (string, error) {
	tokenBytes := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// ResetPassword handles updating the user's password after verifying the reset token.
func ResetPassword(resetToken, newPassword string) error {
	var resetEntry models.PasswordResetToken

	// Find the reset token entry
	if err := config.DB.Where("token = ?", resetToken).First(&resetEntry).Error; err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if the token has expired
	if time.Now().After(resetEntry.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	// Find the user by email
	var user models.User
	if err := config.DB.Where("email = ?", resetEntry.Email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the user's password
	user.Password = string(hashedPassword)

	// Save the updated user password
	if err := config.DB.Save(&user).Error; err != nil {
		return errors.New("failed to update password")
	}

	// Delete the reset token entry after successful password reset
	if err := config.DB.Delete(&resetEntry).Error; err != nil {
		return errors.New("failed to delete reset token")
	}

	return nil
}

// GetUsersPaginated returns paginated users with roles and pagination metadata
func GetUsersPaginated(perPage, page int, sortBy string, sortDesc bool, role string, email string) ([]map[string]interface{}, map[string]interface{}, error) {
	var users []models.UserByAdmin
	var totalRecords int64
	var sortOrder string

	// Hitung offset untuk paginasi
	offset := (page - 1) * perPage

	// Tentukan urutan sort berdasarkan parameter sortDesc
	if sortDesc {
		sortOrder = sortBy + " DESC"
	} else {
		sortOrder = sortBy + " ASC"
	}

	// Inisialisasi query untuk menghitung total records
	query := config.DB.Model(&models.UserByAdmin{}).Where("users.deleted_at IS NULL")

	// Jika parameter role tidak kosong, tambahkan kondisi where untuk role
	if role != "" {
		query = query.Joins("JOIN casbin_rule ON users.username = casbin_rule.v0").Where("casbin_rule.v1 = ? AND casbin_rule.ptype = 'g'", role)
	}

	// Jika parameter email tidak kosong, tambahkan kondisi where untuk email
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}

	// Hitung total jumlah records
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %v", err)
	}

	// Ambil data user dengan paginasi dan sorting
	if err := query.Order(sortOrder).Limit(perPage).Offset(offset).Find(&users).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to fetch users: %v", err)
	}

	// Buat slice untuk menyimpan data pengguna dengan role
	var userResponses []map[string]interface{}

	// Get Casbin enforcer untuk mendapatkan role
	enforcer := helpers.GetCasbinEnforcer()

	// Loop melalui setiap user dan tambahkan field role
	for _, user := range users {
		roles, err := enforcer.GetRolesForUser(user.Username)
		if err != nil || len(roles) == 0 {
			roles = []string{"User"} // Default role jika tidak ada role yang ditemukan
		}

		// Buat map untuk response user dengan role
		userResponse := map[string]interface{}{
			"id":          user.ID,
			"uuid":        user.UUID,
			"username":    user.Username,
			"fullname":    user.Fullname,
			"mobile":      user.Mobile,
			"email":       user.Email,
			"role":        roles[0], // Ambil role pertama (jika ada lebih dari satu)
			"created_at":  user.CreatedAt,
			"updated_at":  user.UpdatedAt,
			"deleted_at":  user.DeletedAt,
			"created_by":  user.CreatedBy,
			"updated_by":  user.UpdatedBy,
			"verified_at": user.VerifiedAt,
		}

		userResponses = append(userResponses, userResponse)
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Buat map untuk informasi paginasi
	paginationData := map[string]interface{}{
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return userResponses, paginationData, nil
}

// GetUserByUUID retrieves user data by UUID using the UserByAdmin model
func GetUserByUUID(userUUID string) (*models.UserByAdmin, error) {
	var user models.UserByAdmin

	// Parse the UUID string into a UUID object
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	// Query the database for the user by UUID using UserByAdmin model
	if err := config.DB.Where("uuid = ?", uuidParsed).Where("deleted_at", nil).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Return the user details
	return &user, nil
}

// AddUserRoleByUUID adds a role to a user by UUID if the role does not already exist in Casbin rules.
func AddUserRoleByUUID(userUUID string, role string) error {
	var user models.UserGetData

	// Parse UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Retrieve the username associated with the given UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).Where("deleted_at", nil).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if the role for the user already exists in the CasbinRule table
	var existingRule models.CasbinRule
	if err := config.DB.Where("ptype = ? AND v0 = ? AND v1 = ?", "g", user.Username, role).First(&existingRule).Error; err == nil {
		// If a record is found, return an error indicating that the role already exists
		return errors.New("role already exists for this user")
	}

	// If the role does not exist, create a new Casbin rule for the user
	newRule := models.CasbinRule{
		Ptype: "g",           // Type "g" means it's a grouping policy (role assignment)
		V0:    user.Username, // V0 is the user (subject)
		V1:    role,          // V1 is the role (object)
	}

	// Add the new rule to the database
	if err := config.DB.Create(&newRule).Error; err != nil {
		return errors.New("failed to assign role to user")
	}

	// Return nil if everything was successful
	return nil
}

func DeleteUserRoleByUUID(userUUID string, roleGuardName string) error {
	var user models.UserGetData

	// Parse UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Retrieve the username associated with the given UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).Where("deleted_at", nil).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if the role_guard_name for the user exists in the CasbinRule table
	var existingRule models.CasbinRule
	if err := config.DB.Where("ptype = ? AND v0 = ? AND v1 = ?", "g", user.Username, roleGuardName).First(&existingRule).Error; err != nil {
		return errors.New("role not found for this user")
	}

	// If the role exists, delete the Casbin rule
	if err := config.DB.Delete(&existingRule).Error; err != nil {
		return errors.New("failed to delete role for user")
	}

	// Return nil if the role was successfully deleted
	return nil
}

// CreateUserByAdmin allows an admin to create a user with a specific role_guard_name
func CreateUserByAdmin(req dto.RegisterRequest, roleGuardName string) (*dto.RegisterResponse, error) {
	var existingUser models.UserRegister

	// Check if the user already exists by username
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create the new user model
	user := models.UserRegister{
		Fullname: req.Fullname,
		Username: req.Username,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "user", "id", "user_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return nil, fmt.Errorf("failed to reset sequence: %w", err)
	}

	// Save the user to the database
	if err := config.DB.Create(&user).Error; err != nil {
		return nil, errors.New("could not create user, please try again")
	}

	// Assign the role from the payload (role_guard_name)
	rule := models.CasbinRule{
		Ptype: "g",
		V0:    req.Username,  // The username (subject)
		V1:    roleGuardName, // The role (object)
	}

	// Save the role to the Casbin rule table
	if err := config.DB.Create(&rule).Error; err != nil {
		return nil, errors.New("user created, but failed to assign role")
	}

	// Create the response
	response := &dto.RegisterResponse{
		ID:       user.ID,
		UUID:     user.UUID,
		Fullname: user.Fullname,
		Username: user.Username,
		Mobile:   user.Mobile,
		Email:    user.Email,
		Message:  "User created and role assigned successfully",
	}

	return response, nil
}

// UpdateUserByAdmin updates the user details by admin and optionally assigns a new role.
func UpdateUserByAdmin(userUUID string, req dto.UpdateUserRequest, roleGuardName string) (*dto.UpdateUserResponse, error) {
	var user models.UserRegister

	// Parse the UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return nil, errors.New("invalid UUID format")
	}

	// Retrieve the user by UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).Where("deleted_at", nil).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Fullname != "" {
		user.Fullname = req.Fullname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Password != "" {
		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	// Save the updated user data
	if err := config.DB.Save(&user).Error; err != nil {
		return nil, errors.New("failed to update user data")
	}

	// If a new role is provided, update the Casbin rule
	if roleGuardName != "" {
		// First, remove the current role (if it exists)
		var existingRule models.CasbinRule
		if err := config.DB.Where("ptype = ? AND v0 = ?", "g", user.Username).First(&existingRule).Error; err == nil {
			config.DB.Delete(&existingRule)
		}

		// Assign the new role
		newRule := models.CasbinRule{
			Ptype: "g",           // Grouping policy
			V0:    user.Username, // The user (subject)
			V1:    roleGuardName, // The new role (object)
		}

		if err := config.DB.Create(&newRule).Error; err != nil {
			return nil, errors.New("failed to assign new role to user")
		}
	}

	// Create the response
	response := &dto.UpdateUserResponse{
		ID:       user.ID,
		UUID:     user.UUID.String(),
		Fullname: user.Fullname,
		Username: user.Username,
		Mobile:   user.Mobile,
		Email:    user.Email,
		Message:  "User updated successfully",
	}

	return response, nil
}

// DeleteUserByAdmin deletes a user and their associated Casbin roles by UUID
func DeleteUserByAdmin(userUUID string) error {
	var user models.UserRegister

	// Parse the UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Retrieve the user by UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).Where("deleted_at", nil).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Delete associated Casbin roles for the user
	if err := config.DB.Where("ptype = ? AND v0 = ?", "g", user.Username).Delete(&models.CasbinRule{}).Error; err != nil {
		return errors.New("failed to delete user roles")
	}

	// Delete the user from the database
	if err := config.DB.Delete(&user).Error; err != nil {
		return errors.New("failed to delete user")
	}

	return nil
}

// ActivateUser activates a user account by setting the VerifiedAt field
func ActivateUser(userUUID string) error {
	var user models.UserRegister

	// Parse the UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Retrieve the user by UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if the user is already verified
	if user.VerifiedAt != nil {
		return errors.New("user is already verified")
	}

	// Set the VerifiedAt field to the current time
	now := time.Now()
	user.VerifiedAt = &now

	// Save the updated user status to the database
	if err := config.DB.Save(&user).Error; err != nil {
		return errors.New("failed to activate user")
	}

	return nil
}

// ActivateUser activates a user account by setting the VerifiedAt field
func DeactivateUser(userUUID string) error {
	var user models.UserRegister

	// Parse the UUID
	uuidParsed, err := uuid.Parse(userUUID)
	if err != nil {
		return errors.New("invalid UUID format")
	}

	// Retrieve the user by UUID
	if err := config.DB.Where("uuid = ?", uuidParsed).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if the user is already verified
	if user.VerifiedAt == nil {
		return errors.New("user is already unverified")
	}

	// Set the VerifiedAt field to the current time
	user.VerifiedAt = nil

	// Save the updated user status to the database
	if err := config.DB.Save(&user).Error; err != nil {
		return errors.New("failed to deactivate user")
	}

	return nil
}
