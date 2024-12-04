package main

import (
	"Project/UX"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"strings"
)

// базовые пути сайта
func basicRoutes(app *fiber.App) {
	basic := app.Group("/")
	basic.Get("/home", UX.Home)
	basic.Post("/login", UX.Login)
	basic.Post("/register", UX.Login)
	app.Post("/generate-key", UX.GenerateKey)
	app.Get("/check-key", UX.CheckKey)
	app.Post("/finalize-registration", UX.FinalizeRegistration)
	app.Post("/finalize-login", UX.FinalizeLogin)
	app.Get("/logout", UX.Logout)

}

// страницы пользователей
func userPages(app *fiber.App) {
	userPages := app.Group("/user")
	userPages.Get("/registration", UX.Reg)
	userPages.Get("/authorization", UX.Auth)
	userPages.Get("/dashboard", UX.Dashboard)
	userPages.Get("/tariffs", UX.Tariffs)
	userPages.Get("/purchases", UX.Purchases)
	userPages.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{"http://localhost:8080", "http://localhost:8080"}, ","),
		AllowCredentials: true,
	}))

	//userPages.Get("/dashboard", UX.Dashboard)

}

func main() {
	app := fiber.New(fiber.Config{})
	//db.Migrations()
	app.Use(logger.New())   // Логирование запросов
	app.Use(compress.New()) // Сжатие ответов
	app.Use(recover.New())  // Восстановление после паники
	// JWT Middlewar
	app.Static("/", "./UI") // подключаем статику

	basicRoutes(app) //базовые маршруты
	userPages(app)
	app.Use(func(c *fiber.Ctx) error { return UX.NotFnd(c) }) //обработчик ошибок
	go func() {
		if err := UX.RunTelegramBot(); err != nil {
			log.Fatalf("Ошибка запуска Telegram бота: %v", err)
		}
	}()

	app.Listen(":8080") //что слушать
}
