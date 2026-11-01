package routers

import (
	"net/http"

	"simple-blackjack/config"
	"simple-blackjack/controller"
	"simple-blackjack/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Basic authentication required"})
			return
		}

		var user models.User
		if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// public routes
	r.POST("/users/register", controller.Register)

	// auth-protected routes
	auth := r.Group("/")
	auth.Use(BasicAuthMiddleware())
	{
		// user & coins
		auth.GET("/users/me", controller.GetProfile)
		auth.POST("/users/coin", controller.AddCoin)

		// blackjack game endpoints
		auth.POST("/games", controller.StartGame)
		auth.GET("/games", controller.ListGames)
		auth.GET("/games/:id", controller.GetGame)
		auth.POST("/games/:id/hit", controller.Hit)
		auth.POST("/games/:id/stand", controller.Stand)
		auth.DELETE("/games/:id", controller.DeleteGame)
	}

	return r
}
