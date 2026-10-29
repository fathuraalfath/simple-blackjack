package controller

import (
	"net/http"

	"simple-blackjack/config"
	"simple-blackjack/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=4"`
	Coin     int    `json:"coin"`
}

type CoinInput struct {
	Amount int `json:"amount" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := config.DB.Where("username = ?", input.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	coin := input.Coin
	if coin < 0 {
		coin = 0
	}

	user := models.User{
		Username: input.Username,
		Password: string(hashed),
		Coin:     coin,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"coin":       user.Coin,
			"created_at": user.CreatedAt,
		},
	})
}

func GetProfile(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user := currentUser.(models.User)

	// Fetch freshest balance from DB
	if err := config.DB.First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"coin":       user.Coin,
		"created_at": user.CreatedAt,
	})
}

func AddCoin(c *gin.Context) {
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input CoinInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	user := currentUser.(models.User)
	if err := config.DB.First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	user.Coin += input.Amount
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update coin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "coin added successfully",
		"added":     input.Amount,
		"user_id":   user.ID,
		"new_total": user.Coin,
	})
}
