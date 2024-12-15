package UX

import (
	"Project/db"
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"html/template"
	"log"
	"os"
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
		ext, user := db.GetUserByName(name)
		if ext {
			if user.Isblocked {
				return true, ""
			} else {
				return false, name
			}
		} else {
			return true, ""
		}
	}
	return true, ""
}

// Функция для отображения Dashboard
func Dashboard(c *fiber.Ctx) error {
	status, username := Restricted(c) //status,username := ...

	if status {
		return c.Redirect("/user/authorization")
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

// Обработчик страницы тарифов
func Purchases(c *fiber.Ctx) error {
	status, username := Restricted(c) // Проверка авторизации

	if status {
		return Auth(c) // Если пользователь не авторизован, перенаправляем
	}

	// Получаем пользователя по имени
	found, _ := db.GetUserUsername(username)
	if !found {
		log.Println("Пользователь не найден")
		return c.Status(fiber.StatusNotFound).SendString("Пользователь не найден")
	}

	// Получаем список тарифов пользователя
	userPlans, err := db.GetUserPlans(username) // Вызов функции получения тарифов
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

	// Передаём данные в шаблон
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

// Обработчик для отправки конфигурации в Telegram
func SendConfig(c *fiber.Ctx) error {
	// Получаем ID тарифа из параметров запроса
	saleID := c.Params("id")
	var sale db.Sale

	// Ищем покупку по ID
	res := db.DB.Where("id = ?", saleID).First(&sale)
	if res.RowsAffected == 0 {
		log.Println("Покупка не найдена")
		return c.Status(fiber.StatusNotFound).SendString("Покупка не найдена")
	}

	// Получаем пользователя, связанного с покупкой
	var user db.User
	res = db.DB.Where("id = ?", sale.Userid).First(&user)
	if res.RowsAffected == 0 {
		log.Println("Пользователь не найден")
		return c.Status(fiber.StatusNotFound).SendString("Пользователь не найден")
	}

	//log.Println(user)
	// Отправка сообщения в Telegram
	message := sale.Config
	//log.Println(message)
	err := SendTelegramConfFile(user.Tgid, "Ваша конфигурация:\n", message)
	if err != nil {
		log.Println("Ошибка отправки сообщения в Telegram:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка отправки сообщения в Telegram")
	}

	return c.SendString("Конфигурация успешно отправлена в Telegram")
}

// Функция отправки текстового файла с расширением .conf в Telegram
func SendTelegramConfFile(tgid uint, fileName string, fileContent string) error {
	// Создаем временный файл с содержимым
	tempFile, err := os.CreateTemp("", "*.conf")
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Удаляем временный файл после отправки

	// Записываем содержимое в файл
	if _, err := tempFile.WriteString(fileContent); err != nil {
		tempFile.Close()
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}
	tempFile.Close()
	//log.Print("tgid")
	//log.Println(tgid)
	// Создаем объект для отправки файла в Telegram
	document := tgbotapi.NewDocument(int64(tgid), tgbotapi.FilePath(tempFile.Name()))
	document.Caption = fileName

	// Отправляем файл через Telegram API
	if _, err := TelegramBot.Send(document); err != nil {
		return fmt.Errorf("ошибка отправки файла: %w", err)
	}

	return nil
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
	//log.Println(payment_method)

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
	//log.Printf("Payment confirmed for product ID: %s\n", productID)

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
	_, user := db.GetUserUsername(username)
	err, sale, _ := db.AddProductToUser(username, productID)
	if !err {
		return c.Status(fiber.StatusBadRequest).SendString("Bad request")
	}
	db.AddLog(strconv.FormatUint(uint64(sale.Id), 10), "Куплен тариф")
	_, conf, _ := AddConf(int(sale.Id))
	SendTelegramConfFile(user.Tgid, "config.conf", conf)
	er := db.AddConfigBySaleID(sale.Id, conf)
	if er != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Bad request")
	}
	// Mock обработка завершения покупки
	return c.Redirect("/user/purchases")
}
