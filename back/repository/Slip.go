package repository

import (
	"PromptPayQR/model"
	"log"
	"gorm.io/gorm"
)

// Add a new slip with transaction reference and image
func CreateSlip(db *gorm.DB, transRef string) model.Slip {
	slip := model.Slip{
		TransRef: transRef,
	}
	result := db.Create(&slip)
	if result.Error != nil {
		log.Fatalf("Failed to create slip: %v", result.Error)
	}
	return slip
}

// Get a slip by transRef
func GetSlipByTransRef(db *gorm.DB, transRef string) (*model.Slip, error) {
	var slip model.Slip
	result := db.Where("trans_ref = ?", transRef).First(&slip)

	// Check if an error occurred
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Slip not found
		}
		return nil, result.Error // Some other error
	}

	// Record found
	return &slip, nil
}
