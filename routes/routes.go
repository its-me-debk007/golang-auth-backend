package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/its-me-debk007/auth-backend/controllers"
)

func SetupRoutes(app *fiber.App) {
	app.Post("/api/login", controllers.Login)
	app.Post("/api/signup", controllers.Signup)
	app.Post("/api/send_otp", controllers.SendOtp)
	app.Patch("/api/reset_password", controllers.ResetPassword)
	app.Post("/api/verify_otp", controllers.VerifyOtp)
	app.Post("/api/refresh", controllers.Refresh)
}
