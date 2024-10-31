package main

import (
	"Project/UX"
	"github.com/gofiber/fiber/v2"
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
	basic.Get("/authorization", UX.Auth)
	basic.Get("/registration", UX.Reg)
	basic.Post("/login", UX.Login)
}

// страницы пользователей
func userPages(app *fiber.App) {
	userPages := app.Group("/user")
	userPages.Get("/dashboard", UX.Dashboard)
	userPages.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{"http://localhost:8080", "http://localhost:8080"}, ","),
		AllowCredentials: true,
	}))
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	app.Use(logger.New())   // Логирование запросов
	app.Use(compress.New()) // Сжатие ответов
	app.Use(recover.New())  // Восстановление после паники
	// JWT Middlewar
	app.Static("/", "./UI") // подключаем статику

	basicRoutes(app) //базовые маршруты
	userPages(app)
	app.Use(func(c *fiber.Ctx) error { return UX.NotFnd(c) }) //обработчик ошибок

	app.Listen(":8080") //что слушать
}
