package UX

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
)

const SecretKey = "secret"

func restricted(c *fiber.Ctx) (bool, string) {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil // Используем SecretKey, который был сгенерирован в функции Login
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return true, ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Status(fiber.StatusUnauthorized)
		return true, ""
	}
	name := claims["name"].(string)
	if name != "" {
		return false, name
	}
	return true, ""
}

// Функция для отображения Dashboard
func Dashboard(c *fiber.Ctx) error {
	status, _ := restricted(c) //status,username := ...
	if status {
		return Auth(c)
	}
	// Загружаем и парсим основной шаблон и шаблон контента
	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/dash.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	// Данные для отображения
	data := map[string]interface{}{
		"TrafficUsed":     "10",
		"ActivePlans":     "2",
		"Promotions":      "8",
		"NextPaymentDate": "15.05",
	}
	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "sidebar", data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Type("html").Send(buf.Bytes())
}
