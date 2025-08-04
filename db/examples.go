package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// ExampleCRUDOperations demonstrates all CRUD operations
func ExampleCRUDOperations() {
	// Initialize CRUD operations
	userCRUD := &UserCRUD{}
	categoryCRUD := &CategoryCRUD{}
	statementCRUD := &StatementCRUD{}

	fmt.Println("=== PostgreSQL CRUD Operations Example ===")

	// USER CRUD OPERATIONS
	fmt.Println("\n--- User CRUD Operations ---")

	// Create users
	user1 := &User{
		ID:       uuid.New().String(),
		Email:    "john@example.com",
		Password: "hashed_password_1",
	}
	user2 := &User{
		ID:       uuid.New().String(),
		Email:    "jane@example.com",
		Password: "hashed_password_2",
	}

	if err := userCRUD.CreateUser(user1); err != nil {
		log.Printf("Error creating user1: %v", err)
	} else {
		fmt.Printf("✓ Created user: %s\n", user1.Email)
	}

	if err := userCRUD.CreateUser(user2); err != nil {
		log.Printf("Error creating user2: %v", err)
	} else {
		fmt.Printf("✓ Created user: %s\n", user2.Email)
	}

	// Read users
	retrievedUser, err := userCRUD.GetUserByEmail("john@example.com")
	if err != nil {
		log.Printf("Error getting user: %v", err)
	} else {
		fmt.Printf("✓ Retrieved user: %s\n", retrievedUser.Email)
	}

	// Update user
	user1.Email = "john.updated@example.com"
	if err := userCRUD.UpdateUser(user1); err != nil {
		log.Printf("Error updating user: %v", err)
	} else {
		fmt.Printf("✓ Updated user email to: %s\n", user1.Email)
	}

	// Get all users
	allUsers, err := userCRUD.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
	} else {
		fmt.Printf("✓ Total users in database: %d\n", len(allUsers))
	}

	// CATEGORY CRUD OPERATIONS
	fmt.Println("\n--- Category CRUD Operations ---")

	// Create categories
	category1 := &Category{
		ID:     uuid.New().String(),
		Title:  "Work",
		UserId: user1.ID,
	}
	category2 := &Category{
		ID:     uuid.New().String(),
		Title:  "Personal",
		UserId: user1.ID,
	}
	category3 := &Category{
		ID:     uuid.New().String(),
		Title:  "Shopping",
		UserId: user2.ID,
	}

	if err := categoryCRUD.CreateCategory(category1); err != nil {
		log.Printf("Error creating category1: %v", err)
	} else {
		fmt.Printf("✓ Created category: %s\n", category1.Title)
	}

	if err := categoryCRUD.CreateCategory(category2); err != nil {
		log.Printf("Error creating category2: %v", err)
	} else {
		fmt.Printf("✓ Created category: %s\n", category2.Title)
	}

	if err := categoryCRUD.CreateCategory(category3); err != nil {
		log.Printf("Error creating category3: %v", err)
	} else {
		fmt.Printf("✓ Created category: %s\n", category3.Title)
	}

	// Read categories
	userCategories, err := categoryCRUD.GetCategoriesByUserID(user1.ID)
	if err != nil {
		log.Printf("Error getting user categories: %v", err)
	} else {
		fmt.Printf("✓ User %s has %d categories:\n", user1.Email, len(userCategories))
		for _, cat := range userCategories {
			fmt.Printf("  - %s\n", cat.Title)
		}
	}

	// Update category
	category1.Title = "Work (Updated)"
	if err := categoryCRUD.UpdateCategory(category1); err != nil {
		log.Printf("Error updating category: %v", err)
	} else {
		fmt.Printf("✓ Updated category title to: %s\n", category1.Title)
	}

	// STATEMENT CRUD OPERATIONS
	fmt.Println("\n--- Statement CRUD Operations ---")

	// Create statements
	statement1 := &Statement{
		ID:         uuid.New().String(),
		Title:      "Complete project documentation",
		UserId:     user1.ID,
		CategoryId: category1.ID,
	}
	statement2 := &Statement{
		ID:         uuid.New().String(),
		Title:      "Review code changes",
		UserId:     user1.ID,
		CategoryId: category1.ID,
	}
	statement3 := &Statement{
		ID:         uuid.New().String(),
		Title:      "Buy groceries",
		UserId:     user2.ID,
		CategoryId: category3.ID,
	}

	if err := statementCRUD.CreateStatement(statement1); err != nil {
		log.Printf("Error creating statement1: %v", err)
	} else {
		fmt.Printf("✓ Created statement: %s\n", statement1.Title)
	}

	if err := statementCRUD.CreateStatement(statement2); err != nil {
		log.Printf("Error creating statement2: %v", err)
	} else {
		fmt.Printf("✓ Created statement: %s\n", statement2.Title)
	}

	if err := statementCRUD.CreateStatement(statement3); err != nil {
		log.Printf("Error creating statement3: %v", err)
	} else {
		fmt.Printf("✓ Created statement: %s\n", statement3.Title)
	}

	// Read statements by category
	categoryStatements, err := statementCRUD.GetStatementsByCategoryID(category1.ID)
	if err != nil {
		log.Printf("Error getting category statements: %v", err)
	} else {
		fmt.Printf("✓ Category '%s' has %d statements:\n", category1.Title, len(categoryStatements))
		for _, stmt := range categoryStatements {
			fmt.Printf("  - %s\n", stmt.Title)
		}
	}

	// Read statements by user
	userStatements, err := statementCRUD.GetStatementsByUserID(user1.ID)
	if err != nil {
		log.Printf("Error getting user statements: %v", err)
	} else {
		fmt.Printf("✓ User %s has %d statements:\n", user1.Email, len(userStatements))
		for _, stmt := range userStatements {
			fmt.Printf("  - %s\n", stmt.Title)
		}
	}

	// Update statement
	statement1.Title = "Complete project documentation (updated)"
	if err := statementCRUD.UpdateStatement(statement1); err != nil {
		log.Printf("Error updating statement: %v", err)
	} else {
		fmt.Printf("✓ Updated statement title to: %s\n", statement1.Title)
	}

	// DELETE OPERATIONS (commented out to preserve data for demonstration)
	fmt.Println("\n--- Delete Operations (commented out) ---")
	fmt.Println("The following delete operations are available but commented out to preserve data:")
	fmt.Println("- DeleteStatement(statementID)")
	fmt.Println("- DeleteCategory(categoryID)")
	fmt.Println("- DeleteUser(userID)")
	fmt.Println("- DeleteStatementsByUserID(userID)")
	fmt.Println("- DeleteStatementsByCategoryID(categoryID)")
	fmt.Println("- DeleteCategoriesByUserID(userID)")

	// Example of how to delete (uncomment to use):
	/*
		if err := statementCRUD.DeleteStatement(statement1.ID); err != nil {
			log.Printf("Error deleting statement: %v", err)
		} else {
			fmt.Printf("✓ Deleted statement: %s\n", statement1.Title)
		}
	*/

	fmt.Println("\n=== CRUD Operations Example Completed ===")
}
