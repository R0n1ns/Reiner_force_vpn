package main

import (
	"Project/UX"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"html/template"
	"log"
)

func notFnd(c *fiber.Ctx) error {
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./UI/notFaund.gohtml")
	if err != nil {
		log.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, c); err != nil {
		log.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Отправляем отрендеренный шаблон в ответ
	return c.Type("html").Send(buf.Bytes())
}

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
	// Подключаем middleware
	app.Use(logger.New())   // Логирование запросов
	app.Use(compress.New()) // Сжатие ответов
	app.Use(recover.New())  // Восстановление после паники
	app.Static("/", "./UI")

	//базовые маршруты
	basicRoutes(app)
	app.Use(func(c *fiber.Ctx) error {
		// Если маршрут не найден, вызываем функцию notFnd
		return notFnd(c)
	})
	//RegisterProductRoutes(app)
	app.Listen(":8080")
}
