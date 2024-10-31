package UX

import (
	"Project/db"
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
	"time"
)

// ----------------------- регистрация -----------------------
// страница регистрации
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
func RegNew(data *fiber.Ctx) {

}

// ----------------------- авторизация -----------------------
// страница авторизации

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

func Login(c *fiber.Ctx) error {
	var dt map[string]string
	json.Unmarshal(c.Body(), &dt)
	user := dt["email"]
	pass := dt["password"]

	// Throws Unauthorized error
	d, data := db.GetUser(user)
	if !d || data.Password != pass {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name": data.UserName,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	}
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    t,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
