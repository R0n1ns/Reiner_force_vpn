package main

import (
	"Project/Handlers"
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
	basic.Get("/home", Handlers.Home)
	basic.Post("/login", Handlers.Login)
	basic.Post("/register", Handlers.Login)
	app.Post("/generate-key", Handlers.GenerateKey)
	app.Get("/check-key", Handlers.CheckKey)
	app.Post("/finalize-registration", Handlers.FinalizeRegistration)
	app.Post("/finalize-login", Handlers.FinalizeLogin)
	app.Get("/logout", Handlers.Logout)
	app.Get("/support", Handlers.FAQ)

}

// страницы пользователей
func userPages(app *fiber.App) {
	userPages := app.Group("/user")
	userPages.Get("/registration", Handlers.Reg)
	userPages.Get("/authorization", Handlers.Auth)
	userPages.Get("/dashboard", Handlers.Dashboard)
	userPages.Get("/tariffs", Handlers.Tariffs)
	userPages.Get("/purchases", Handlers.Purchases)
	userPages.Get("/tariff/:id", Handlers.PaymentPage)
	// Route to redirect to payment
	userPages.Get("/redirect-payment", Handlers.RedirectToPayment)
	// Route to confirm payment
	userPages.Post("/confirm-payment", Handlers.ConfirmPayment)
	// Route to finalize the purchase
	userPages.Get("/sale", Handlers.FinalizeSale)
	userPages.Post("/send-config/:id", Handlers.SendConfig)

	userPages.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{"http://localhost:8080", "http://localhost:8080"}, ","),
		AllowCredentials: true,
	}))

	//userPages.Get("/dashboard", Handlers.Dashboard)

}

// страницы пользователей
func adminPages(app *fiber.App) {
	adminPages := app.Group("/admin")
	adminPages.Get("/dashboard", Handlers.AdminDashboard)
	adminPages.Get("/userspanel", Handlers.UsersPanel)
	adminPages.Post("/blockuser", Handlers.Blockuser)
	adminPages.Post("/deleteuser", Handlers.DeleteUser)
	adminPages.Get("/logs", Handlers.Logs)
	adminPages.Get("/products", Handlers.Products)
	adminPages.Get("products/add", Handlers.AddProductPage)        // Страница добавления продукта
	adminPages.Post("products/saveadd", Handlers.AddProduct)       // Обработка добавления продукта
	adminPages.Get("products/edit/:id", Handlers.EditProductPage)  // Страница редактирования продукта
	adminPages.Post("products/saveedit/:id", Handlers.EditProduct) // Обработка редактирования продукта
	adminPages.Get("products/delete/:id", Handlers.DeleteProduct)  // Удаление продукта

	adminPages.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join([]string{"http://localhost:8080", "http://localhost:8080"}, ","),
		AllowCredentials: true,
	}))

	//userPages.Get("/dashboard", Handlers.Dashboard)

}

func main() {
	defer func() {
		err := Handlers.Wg_client.SaveToFile(Handlers.Filename)
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
	app.Static("/", "./Templates") // подключаем статику
	basicRoutes(app)               //базовые маршруты
	userPages(app)
	adminPages(app)
	app.Use(func(c *fiber.Ctx) error { return Handlers.NotFnd(c) }) //обработчик ошибок
	go func() {
		if err := Handlers.RunTelegramBot(); err != nil {
			log.Fatalf("Ошибка запуска Telegram бота: %v", err)
		}
	}()
	go func() {
		// Запуск планировщика
		s := gocron.NewScheduler(time.UTC)
		s.Every(1).Minutes().Do(func() {
			Handlers.UpdateTraffic()
		})
		s.Every(1).Days().Do(func() {
			Handlers.DeleteExpiredSales()
		})
		s.StartBlocking()
	}()
	//go func() {
	//	Handlers.ScheduleDeletion()
	//}()
	app.Listen(":8080") //что слушать
}
