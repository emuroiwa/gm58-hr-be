package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"gm58-hr-backend/internal/config"
	"gm58-hr-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main111() {
	var (
		username = flag.String("username", "", "Super admin username")
		email    = flag.String("email", "", "Super admin email")
		password = flag.String("password", "", "Super admin password")
	)
	flag.Parse()

	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Get username
	if *username == "" {
		fmt.Print("Enter super admin username: ")
		*username, _ = reader.ReadString('\n')
		*username = strings.TrimSpace(*username)
	}

	// Get email
	if *email == "" {
		fmt.Print("Enter super admin email: ")
		*email, _ = reader.ReadString('\n')
		*email = strings.TrimSpace(*email)
	}

	// Get password
	if *password == "" {
		fmt.Print("Enter super admin password: ")
		bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
		*password = string(bytePassword)
		fmt.Println()
	}

	// Validate inputs
	if *username == "" || *email == "" || *password == "" {
		log.Fatal("Username, email, and password are required")
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", *username, *email).First(&existingUser).Error; err == nil {
		fmt.Println("User already exists. Updating to super admin...")

		// Update existing user to super admin
		existingUser.Role = "super_admin"
		if err := db.Save(&existingUser).Error; err != nil {
			log.Fatal("Failed to update user:", err)
		}

		fmt.Println("✅ User updated to super admin successfully!")
		fmt.Printf("Username: %s\n", existingUser.Username)
		fmt.Printf("Email: %s\n", existingUser.Email)
		fmt.Printf("Role: %s\n", existingUser.Role)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create super admin user
	superAdmin := models.User{
		Username: *username,
		Email:    *email,
		Password: string(hashedPassword),
		Role:     "super_admin",
		IsActive: true,
	}

	if err := db.Create(&superAdmin).Error; err != nil {
		log.Fatal("Failed to create super admin:", err)
	}

	fmt.Println("✅ Super admin created successfully!")
	fmt.Printf("Username: %s\n", superAdmin.Username)
	fmt.Printf("Email: %s\n", superAdmin.Email)
	fmt.Printf("Role: %s\n", superAdmin.Role)
	fmt.Println("\nYou can now login with these credentials and access all companies.")
}
