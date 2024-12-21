package Handlers

import (
	"Project/db"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"html/template"
	"log"
	"strconv"
	"time"
)

const AdminSecretKey = "secret"

// Проверка авторизации администратора
func AdminRestricted(c *fiber.Ctx) (bool, string) {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(AdminSecretKey), nil
	})

	if err != nil {
		return true, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true, ""
	}
	is_admin := claims["is_admin"].(bool)
	if !is_admin {
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

// Панель администратора
func AdminDashboard(c *fiber.Ctx) error {
	status, adminName := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	// Подсчет новых данных за последние сутки
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	newUsersCount := db.CountNewUsers(yesterday, now)
	newPurchasesCount := db.CountNewPurchases(yesterday, now)
	newLogsCount := db.CountLogs(yesterday, now)

	// Данные для отображения
	dashboardData := map[string]interface{}{
		"AdminName":         adminName,
		"NewUsersCount":     newUsersCount,
		"NewPurchasesCount": newPurchasesCount,
		"NewLogsCount":      newLogsCount,
	}

	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/dashadmin.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", dashboardData); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Управление пользователями
func UsersPanel(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	// Загрузка шаблонов
	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/userspanel.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Получение списка пользователей
	users := db.GetUsers()
	if users == nil {
		log.Println("Ошибка получения данных пользователей")
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Подготовка данных для отображения
	var buf bytes.Buffer
	//data := map[string]interface{}{
	//	"Users": users,
	//}

	// Рендеринг шаблона
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", users); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}
func Blockuser(c *fiber.Ctx) error {
	userID := c.FormValue("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	err, user := db.ToggleBlockUser(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error blocking user")
	}
	if user.Isblocked {
		db.AddLog(userID, "Блокировка пользователя")
	} else {
		db.AddLog(userID, "Разблокировка пользователя")

	}
	return c.Redirect("/admin/userspanel")
}
func DeleteUser(c *fiber.Ctx) error {
	userID := c.FormValue("id")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	if err := db.DeleteUserById(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error deleting user")
	}
	db.AddLog(userID, "Удаление пользователя")
	return c.Redirect("/admin/userspanel")
}

// // Логи системы
func Logs(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/logs.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	logs := db.GetSystemLogs()

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", map[string]interface{}{
		"Logs": logs,
	}); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Управление продуктами
func Products(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/products.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	products := db.Getproducts()
	if products == nil {
		log.Println("Ошибка получения продуктов:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", map[string]interface{}{
		"Products": products,
	}); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Обработка страницы добавления продукта
func AddProductPage(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/add_product.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", nil); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

func AddProduct(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	// Создаем продукт с общими данными
	product := db.Product{
		Name:      c.FormValue("Name"),
		NowPrice:  parseUint(c.FormValue("NowPrice"), 0),
		LastPrice: parseUint(c.FormValue("LastPrice"), 0),
		IsOnSale:  c.FormValue("IsOnSale") == "on",
	}

	// Определяем тип тарифа
	tariffType := c.FormValue("TariffType")
	if tariffType == "term" {
		product.IsTerm = true
		product.Term = parseUint(c.FormValue("Term"), 0)
	} else if tariffType == "traffic" {
		product.IsTraffic = true
		product.Traffic = parseUint(c.FormValue("Traffic"), 0)
	} else {
		log.Println("Некорректный тип тарифа")
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Tariff Type")
	}

	// Сохранение продукта в базу данных
	if err := db.DB.Create(&product).Error; err != nil {
		log.Println("Ошибка добавления продукта в базу данных:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	//log.Println("Продукт успешно добавлен:", product)
	db.AddLog(product.Name, "Продукт добавлен")

	return c.Redirect("/admin/products")
}

func parseUint(value string, fallback uint) uint {
	if parsed, err := strconv.ParseUint(value, 10, 32); err == nil {
		return uint(parsed)
	}
	return fallback
}

func EditProductPage(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	id := c.Params("id")
	product := new(db.Product)
	if err := db.DB.First(product, id).Error; err != nil {
		log.Println("Продукт не найден:", err)
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}

	tmpl, err := template.ParseFiles("./Templates/sidebaradmin.gohtml", "./Templates/edit_product.gohtml")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "sidebar", product); err != nil {
		log.Println("Ошибка выполнения шаблона:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	return c.Type("html").Send(buf.Bytes())
}

// Обработка редактирования продукта
func EditProduct(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	id := c.Params("id")
	product := new(db.Product)
	if err := db.DB.First(product, id).Error; err != nil {
		log.Println("Продукт не найден:", err)
		return c.Status(fiber.StatusNotFound).SendString("Product not found")
	}

	// Обновляем поля продукта
	product.Name = c.FormValue("Name")
	product.NowPrice = parseUint(c.FormValue("NowPrice"), product.NowPrice)
	product.LastPrice = parseUint(c.FormValue("LastPrice"), product.LastPrice)
	product.IsOnSale = c.FormValue("IsOnSale") == "on"

	tariffType := c.FormValue("TariffType")
	if tariffType == "term" {
		product.IsTerm = true
		product.Term = parseUint(c.FormValue("Term"), 0)
		product.IsTraffic = false
		product.Traffic = 0
	} else if tariffType == "traffic" {
		product.IsTraffic = true
		product.Traffic = parseUint(c.FormValue("Traffic"), 0)
		product.IsTerm = false
		product.Term = 0
	}

	if err := db.DB.Save(product).Error; err != nil {
		log.Println("Ошибка обновления продукта в базе данных:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	db.AddLog(product.Name, "Продукт изменен")

	return c.Redirect("/admin/products")
}

func DeleteProduct(c *fiber.Ctx) error {
	status, _ := AdminRestricted(c)
	if status {
		return c.Redirect("/login")
	}

	// Получаем ID продукта из параметров URL
	id := c.Params("id")

	// Удаление продукта и связанных с ним продаж
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Удаление продаж, связанных с этим продуктом
		if err := tx.Where("productid = ?", id).Delete(&db.Sale{}).Error; err != nil {
			log.Println("Ошибка при удалении продаж:", err)
			return err
		}

		// Удаление самого продукта
		if err := tx.Delete(&db.Product{}, id).Error; err != nil {
			log.Println("Ошибка при удалении продукта:", err)
			return err
		}

		return nil
	}); err != nil {
		log.Println("Ошибка транзакции при удалении продукта:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	db.AddLog(id, "Удаление продукта")

	return c.Redirect("/admin/products")
}
