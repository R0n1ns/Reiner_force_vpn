package main

import (
	"Project/UX"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"strings"
	"time"
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
	app.Get("/support", UX.FAQ)

}

// страницы пользователей
func userPages(app *fiber.App) {
	userPages := app.Group("/user")
	userPages.Get("/registration", UX.Reg)
	userPages.Get("/authorization", UX.Auth)
	userPages.Get("/dashboard", UX.Dashboard)
	userPages.Get("/tariffs", UX.Tariffs)
	userPages.Get("/purchases", UX.Purchases)
	userPages.Get("/tariff/:id", UX.PaymentPage)
	// Route to redirect to payment
	userPages.Get("/redirect-payment", UX.RedirectToPayment)
	// Route to confirm payment
	userPages.Post("/confirm-payment", UX.ConfirmPayment)
	// Route to finalize the purchase
	userPages.Get("/sale", UX.FinalizeSale)
	userPages.Post("/send-config/:id", UX.SendConfig)

	userPages.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{"http://localhost:8080", "http://localhost:8080"}, ","),
		AllowCredentials: true,
	}))

	//userPages.Get("/dashboard", UX.Dashboard)

}

func main() {
	defer func() {
		err := UX.Wg_client.SaveToFile(UX.Filename)
		if err != nil {
			fmt.Printf("Ошибка при сохранении файла: %v", err)
		} else {
			fmt.Println("Конфигурация успешно сохранена.")
		}
	}()
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
	go func() {
		// Запуск планировщика
		s := gocron.NewScheduler(time.UTC)
		s.Every(1).Hour().Do(func() {
			UX.UpdateTraffic()
		})
		s.StartBlocking()
	}()
	app.Listen(":8080") //что слушать
}
