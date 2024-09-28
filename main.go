package main

import (
	"Project/UX"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// базовые пути сайта
func basicRoutes(app *fiber.App) {
	basic := app.Group("/")
	basic.Get("/home", UX.Home)
	basic.Get("/authorization", UX.Auth)
	basic.Get("/registration", UX.Reg)
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	app.Use(logger.New())   // Логирование запросов
	app.Use(compress.New()) // Сжатие ответов
	app.Use(recover.New())  // Восстановление после паники
	app.Static("/", "./UI") // подключаем статику

	basicRoutes(app) //базовые маршруты

	app.Use(func(c *fiber.Ctx) error { return UX.NotFnd(c) }) //обработчик ошибок

	app.Listen(":8080") //что слушать
}

//Pp_1234567
