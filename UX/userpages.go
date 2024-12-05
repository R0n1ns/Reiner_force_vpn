package UX

import (
	"Project/db"
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
	"strconv"
)

const SecretKey = "secret"

func Restricted(c *fiber.Ctx) (bool, string) {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil // Используем SecretKey, который был сгенерирован в функции Login
	})

	if err != nil {
		return true, ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
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
	status, username := Restricted(c) //status,username := ...

	if status {
		return Auth(c)
	}
	// Загружаем и парсим основной шаблон и шаблон контента
	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/dash.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	totalTraffic, latestExpiration, activatedCount, err := db.GetUserStatistics(username)
	if err != nil {
		fmt.Println("Ошибка:", err)
	}
	var formattedExpiration string
	if !latestExpiration.IsZero() {
		formattedExpiration = latestExpiration.Format("02.01.2006") // Формат: день.месяц.год
	}
	promo, _ := db.CountProductsOnSale()
	// Данные для отображения
	data := map[string]interface{}{
		"TrafficUsed":     totalTraffic,
		"ActivePlans":     activatedCount,
		"Promotions":      promo,
		"NextPaymentDate": formattedExpiration,
	}
	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "sidebar", data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Type("html").Send(buf.Bytes())
}

// Функция для отображения Dashboard
func Tariffs(c *fiber.Ctx) error {
	status, _ := Restricted(c) //status,username := ...

	if status {
		return Auth(c)
	}
	// Загружаем и парсим основной шаблон и шаблон контента
	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/tarifs.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	// Получаем данные продуктов из базы данных
	products := db.Getproducts()
	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "sidebar", *products); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Type("html").Send(buf.Bytes())
}

func Purchases(c *fiber.Ctx) error {
	status, username := Restricted(c) // Проверка авторизации

	if status {
		return Auth(c) // Если пользователь не авторизован, перенаправляем
	}

	// Получаем список тарифов пользователя
	userPlans, err := db.GetUserPlans(username)
	if err != nil {
		log.Println("Ошибка получения тарифов:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Загружаем и парсим основной шаблон и шаблон контента
	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/purchases.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Если тарифов нет, передаем пустой список
	data := map[string]interface{}{
		"UserPlans": userPlans,
	}

	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "sidebar", data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Функция для отображения FAQ
func FAQ(c *fiber.Ctx) error {
	status, _ := Restricted(c) // Проверка авторизации

	if status {
		return Auth(c)
	}
	// Загружаем и парсим основной шаблон (sidebar) и FAQ шаблон
	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/faq.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	// Данные для FAQ (если потребуется передача из сервера)
	data := map[string]interface{}{}
	// Рендерим шаблон в буфер
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "sidebar", data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	return c.Type("html").Send(buf.Bytes())

}
func PaymentPage(c *fiber.Ctx) error {
	// Получаем ID товара из параметров URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		log.Println("Неверный ID продукта:", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Product ID")
	}

	// Ищем товар в базе данных
	ex, product := db.GetProductId(id)
	if !ex {
		log.Println("Ошибка поиска продукта:", err)
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}

	// Передаём данные в шаблон
	data := map[string]interface{}{
		"Id":        product.Id,
		"Name":      product.Name,
		"NowPrice":  product.NowPrice,
		"LastPrice": product.LastPrice,
		"IsOnSale":  product.IsOnSale,
		"IsTerm":    product.IsTerm,
		"Term":      product.Term,
		"IsTraffic": product.IsTraffic,
		"Traffic":   product.Traffic,
	}

	tmpl, err := template.ParseFiles("./UI/sidebar.gohtml", "./UI/payment.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", data); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Обработчик редиректа на платёжный сервис
func RedirectToPayment(c *fiber.Ctx) error {
	productID := c.Query("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Product ID is required")
	}
	payment_method := c.FormValue("payment_method")
	paymentLink := "https://example.com/payment?product_id=" + productID

	if payment_method == "test" {
		paymentLink = "https://google.com"
	} else if payment_method == "ymoney" {
		paymentLink = "https://example.com/payment?product_id=" + productID
	}
	log.Println(payment_method)

	// Генерация ссылки оплаты
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"payment_link": paymentLink})
}

// Обработчик подтверждения оплаты
func ConfirmPayment(c *fiber.Ctx) error {
	productID := c.FormValue("product_id")

	if productID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Product ID is required")
	}
	// Mock проверка оплаты
	log.Printf("Payment confirmed for product ID: %s\n", productID)

	// Перенаправление на страницу завершения покупки
	return c.Redirect("/user/sale?product_id=" + productID)
}

// Обработчик завершения покупки
func FinalizeSale(c *fiber.Ctx) error {
	_, username := Restricted(c) // Проверка авторизации

	productID := c.Query("product_id")
	if productID == "" || username == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Product ID and username are required")
	}
	err, _, _ := db.AddProductToUser(username, productID)
	if !err {
		return c.Status(fiber.StatusBadRequest).SendString("Bad request")
	}

	// Mock обработка завершения покупки
	log.Printf("Sale finalized for product ID: %s by user: %s\n", productID, username)

	return c.Redirect("/user/purchases")
}
