package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"email-job/internal/emails"
	"email-job/internal/middleware"
)

func NewUserRouter(
	emailHandler    *emails.EmailHandler,
	authMiddleware  *middleware.AuthMiddleware,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	v1 := r.Group("/api/v1")
	{
		email := v1.Group("/emails")
		{
			email.POST("/contact",             emailHandler.SendContactMessage)

			email.GET("/live-demo",            authMiddleware.AdminSecretMiddleware(), emailHandler.GetEmailsByLiveDemoRequest)
			email.POST("/live-demo",           emailHandler.SendLiveDemoRequest)
			email.POST("/live-demo/:id/ready", authMiddleware.AdminSecretMiddleware(),emailHandler.SendLiveDemoReady)

			email.GET("id/:id",   authMiddleware.AdminSecretMiddleware(), emailHandler.GetEmailByID)
			email.GET("ip/:ip",   authMiddleware.AdminSecretMiddleware(), emailHandler.GetEmailByIP)
			email.GET("",       authMiddleware.AdminSecretMiddleware(), emailHandler.GetEmails)
			email.GET("/today", authMiddleware.AdminSecretMiddleware(), emailHandler.GetTodayEmails)
		}
	}

	return r
}