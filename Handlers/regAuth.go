package Handlers

import (
	"Project/db"
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	telegramBotToken = "7767402806:AAFpWS_2UFWfOFri6sG6fva4bKxh-cs-Jos"
	BotUsername      string
)
var TelegramBot = &tgbotapi.BotAPI{}

var (
	temporaryKeys       = make(map[string]int64) // Хранилище ключей подтверждения и их tgid
	temporaryKeysStatus = make(map[string]bool)  // Хранилище статуса подтверждения ключей
)

// ----------------------- Telegram Bot -----------------------

func RunTelegramBot() error {
	var err error
	TelegramBot, err = tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Printf("Ошибка авторизации Telegram бота: %v", err)
		return err
	}

	BotUsername = TelegramBot.Self.UserName
	log.Printf("Telegram бот авторизован как %s", BotUsername)

	updates := TelegramBot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message != nil && update.Message.Command() == "start" {
			handleTelegramStart(TelegramBot, update)
		}
	}
	return nil
}
func handleTelegramStart(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update) {
	key := update.Message.CommandArguments()
	tgid := update.Message.Chat.ID

	if _, exists := temporaryKeys[key]; !exists {
		tgbotapi.NewMessage(tgid, "Ошибка: неверный или истёкший ключ.")
		return
	}

	temporaryKeys[key] = tgid
	temporaryKeysStatus[key] = true

	msg := tgbotapi.NewMessage(tgid, "Код Подтвержден!")
	telegramBot.Send(msg)
}

// ----------------------- регистрация -----------------------
// страница регистрации
func Reg(c *fiber.Ctx) error {
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./Templates/registr.gohtml")
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

//func RegNew(data *fiber.Ctx) {
//
//}

func GenerateKey(c *fiber.Ctx) error {

	// Генерация уникального ключа для подтверждения
	key := fmt.Sprintf("%d", time.Now().UnixNano())
	// Временное сохранение ключа и статуса
	temporaryKeys[key] = 0 // Телеграм ID будет обновлён при подтверждении
	temporaryKeysStatus[key] = false

	// Генерация ссылки на Telegram с ключом
	telegramLink := fmt.Sprintf("https://t.me/%s?start=%s", BotUsername, key)

	return c.JSON(fiber.Map{"success": true, "key": key, "telegram_link": telegramLink})
}

func CheckKey(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ключ не указан"})
	}
	fmt.Println(temporaryKeys)
	// Проверка наличия ключа
	tgid, exists := temporaryKeys[key]
	if !exists {
		return c.JSON(fiber.Map{"confirmed": false, "message": "Ключ не найден"})
	}

	// Проверка статуса подтверждения
	if temporaryKeysStatus[key] {
		return c.JSON(fiber.Map{"confirmed": true, "tgid": tgid})
	}

	return c.JSON(fiber.Map{"confirmed": false})
}

func FinalizeRegistration(c *fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Key      string `json:"key"`
	}

	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Неверный формат запроса"})
	}
	// Проверка ключа
	tgid, exists := temporaryKeys[request.Key]
	if !exists || !temporaryKeysStatus[request.Key] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ключ не подтверждён"})
	}
	// Извлекаем имя пользователя из email
	username := strings.Split(request.Email, "@")[0]

	ex, _ := db.GetUserUsername(username)
	if ex {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Пользователь уже существует"})
	}
	ex, _ = db.GetUserByTelegramID(tgid)
	if ex {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Пользователь уже существует"})
	}
	msg := tgbotapi.NewMessage(tgid, "Регистрация успешно подтверждена!\nПерейдите обратно на страницу.")
	TelegramBot.Send(msg)

	// Сохраняем пользователя в базу данных
	user := db.User{
		Mail:     request.Email,
		Password: request.Password,
		UserName: username,
		Tgid:     uint(tgid),
	}
	db.Adduser(user)

	// Удаление временных данных
	delete(temporaryKeys, request.Key)
	delete(temporaryKeysStatus, request.Key)

	db.AddLog(username, "Регистрация пользователя")

	// Создаём JWT Claims
	claims := jwt.MapClaims{
		"name":  username,
		"email": request.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Токен действует 72 часа
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Ошибка при создании токена"})
	}

	// Устанавливаем JWT токен в cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    t,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	// Отправляем успешный ответ с токеном
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Регистрация завершена",
		"token":   t,
	})
}

// ----------------------- авторизация --W---------------------
// страница авторизации

func Auth(c *fiber.Ctx) error {
	status, _ := Restricted(c)
	if !status {
		cookie := c.Cookies("jwt")

		token, _ := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(AdminSecretKey), nil
		})

		claims := token.Claims.(jwt.MapClaims)
		is_admin := claims["is_admin"].(bool)
		if is_admin {
			return c.Redirect("/admin/dashboard")
		} else {
			return c.Redirect("/user/dashboard")
		}
	}
	// Парсим файл шаблона
	tmpl, err := template.ParseFiles("./Templates/auth.gohtml")
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
		"name":     data.UserName,
		"is_admin": data.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
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

	if data.IsAdmin {
		//return c.Redirect("/admin/dashboard")
		return c.JSON(fiber.Map{
			"message": "success", "typ": "admin",
		})
	} else {
		//return c.Redirect("/user/dashboard")
		return c.JSON(fiber.Map{
			"message": "success", "typ": "user",
		})
	}

}

func Logout(c *fiber.Ctx) error {
	// Создаем cookie с пустым значением и истекшим сроком действия
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Срок действия в прошлом
		HTTPOnly: true,
	}
	// Удаляем cookie
	c.Cookie(&cookie)
	// Перенаправляем на страницу /login
	return c.Redirect("/user/authorization")
}

func FinalizeLogin(c *fiber.Ctx) error {
	var request struct {
		Key string `json:"key"`
	}
	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Неверный формат запроса"})
	}
	// Проверка ключа
	tgid, exists := temporaryKeys[request.Key]
	if !exists || !temporaryKeysStatus[request.Key] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ключ не подтверждён"})
	}

	// Throws Unauthorized error
	d, data := db.GetUserByTelegramID(tgid)

	if !d {
		fmt.Println("логин не пройден")
		//msg := tgbotapi.NewMessage(tgid, "Аккаунта с таким id не существует :(")
		if _, err := TelegramBot.Send(tgbotapi.NewMessage(tgid, "Аккаунта с таким id не существует :(")); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Аккаунта с таким id не существует"})
	}
	//fmt.Println("логин пройден")

	msg := tgbotapi.NewMessage(tgid, "Успешное подтверждение авторизации!\nПерейдите обратно на страницу.")
	if _, err := TelegramBot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	// Create the Claims
	claims := jwt.MapClaims{
		"name":     data.UserName,
		"is_admin": data.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Удаление временных данных
	delete(temporaryKeys, request.Key)
	delete(temporaryKeysStatus, request.Key)
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
	// Отправляем успешный ответ с токеном
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Регистрация завершена",
		"token":   t,
	})
}
