package UX

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"html/template"
	"log"
)

func Reg(c *fiber.Ctx) error {
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./UI/registr.gohtml")
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

func Auth(c *fiber.Ctx) error {
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./UI/auth.gohtml")
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
