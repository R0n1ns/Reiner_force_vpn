package UX

import (
	"Project/db"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"html/template"
	"log"
)

func Home(c *fiber.Ctx) error {
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./UI/mainPage.gohtml")
	if err != nil {
		log.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Получаем данные продуктов из базы данных
	products := db.Getproducts()

	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, *products); err != nil {
		log.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Отправляем отрендеренный шаблон в ответ
	return c.Type("html").Send(buf.Bytes())
}
