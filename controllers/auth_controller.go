package controllers

import (
	"EventPlanner_Backend/config"
	"EventPlanner_Backend/models"
	"EventPlanner_Backend/utils"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Message string `json:"message"`
	UserID  uint   `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Token   string `json:"token"`
}

func Signup(c *gin.Context) {
	var input SignupInput

	// ADD THIS - Log the raw request body
	body, _ := c.GetRawData()
	fmt.Printf("=== RAW REQUEST BODY ===\n")
	fmt.Printf("%s\n", string(body))
	fmt.Printf("========================\n")

	// Reset the body so Gin can bind it
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&input); err != nil {
		// ADD DETAILED ERROR LOGGING
		fmt.Printf("=== VALIDATION ERROR ===\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Received data - Name: '%s', Email: '%s', Password length: %d\n",
			input.Name, input.Email, len(input.Password))
		fmt.Printf("========================\n")

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("=== VALIDATION PASSED ===\n")
	fmt.Printf("Name: '%s', Email: '%s', Password: '%s'\n", input.Name, input.Email, input.Password)
	fmt.Printf("========================\n")

	// Rest of your existing signup logic...
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "هذا البريد الإلكتروني مسجل بالفعل"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في تشفير كلمة المرور"})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء الحساب"})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء التوكن"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "تم إنشاء الحساب بنجاح",
		UserID:  user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Token:   token,
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "البريد الإلكتروني أو كلمة المرور غير صحيحة"})
		return
	}

	if !utils.CheckPassword(user.Password, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "البريد الإلكتروني أو كلمة المرور غير صحيحة"})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "فشل في إنشاء التوكن"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Message: "تم تسجيل الدخول بنجاح",
		UserID:  user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Token:   token,
	})
}
