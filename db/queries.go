package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

func init_() *gorm.DB {
	var load_err = godotenv.Load("config.env")
	if load_err != nil {
		log.Println("Ошибка подключения к базе", load_err)
	}

	var db_path = os.Getenv("DB_PATH")

	var db, err = gorm.Open(postgres.Open(db_path), &gorm.Config{})
	if err != nil {
		log.Println("Ошибка подключения к базе", err)
	}
	return db
}

var db = init_()
var DB = db

// миграция баз данных
func Migrations() {
	//снос таблиц
	//db.Migrator().DropTable(&User{})
	//db.Migrator().DropTable(&Products{})
	db.Migrator().DropTable(&Sale{})

	//выполнение миграций
	//db.AutoMigrate(&User{})
	//db.AutoMigrate(&Product{})
	db.AutoMigrate(&Sale{})

	//отчет
	log.Println("Миграции выполнены")
}

// -------------------------------- ПОЛЬЗОВАТЕЛИ --------------------------------

// добовление нового пользователя
func Adduser(user User) bool {
	result := db.Create(&user) // pass pointer of data to Create
	if result.Error == nil {
		//log.Println("Добавлен пользователь ", user.ID)
		return true
	} else {
		log.Println("Ошибка добавдение данных", result.Error)
		return false
	}
}

// получить всех пользователей
func GetUsers() *[]User {
	var users []User
	err := db.Find(&users)
	if err.Error != nil {
		log.Println(err)
		return &[]User{}
	}
	return &users
}

// получение пользователя
func GetUser(mail string) (bool, User) {
	var user User
	res := db.Find(&user, "Mail = ?", mail)
	if res.Error == nil {
		return true, user
	} else {
		return false, User{}
	}
}

func GetUserPlans(username string) ([]map[string]interface{}, error) {
	var user User
	if err := db.Where("user_name = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	var sales []Sale
	if err := db.Preload("Product").Where("userid = ?", user.Id).Find(&sales).Error; err != nil {
		return nil, err
	}

	// Преобразуем данные в формат для шаблона
	userPlans := make([]map[string]interface{}, 0)
	for _, sale := range sales {
		plan := map[string]interface{}{
			"Id":               sale.Id,
			"PlanName":         sale.Product.Name,
			"Status":           "Активен", // Логика проверки статуса
			"RemainingTraffic": sale.RemainingTraffic,
			"ExpirationDate":   sale.ExpirationDate.Format("02.01.2006"),
			"IsRenewable":      !sale.ISFrozen,
			"Config":           sale.Config, // Добавляем конфигурацию
		}
		userPlans = append(userPlans, plan)
	}

	return userPlans, nil
}

// получение пользователя
func GetUserUsername(usrname string) (bool, User) {
	var user User
	res := db.Where("user_name = ?", usrname).First(&user)
	if res.RowsAffected > 0 {
		return true, user
	}
	return false, User{}
}

func GetUserByTelegramID(tgid int64) (bool, User) {
	var user User
	res := db.Where("tgid = ?", tgid).First(&user)
	if res.RowsAffected > 0 {
		return true, user
	}
	return false, User{}
}

// пполучение данных пользователя
// изменения пользователя
// изменения пароля пользователя
// Функция добавления Config по ID продажи
func AddConfigBySaleID(saleID uint, newConfig string) error {
	var sale Sale

	// Находим запись продажи по ID
	if err := db.First(&sale, saleID).Error; err != nil {
		return fmt.Errorf("не удалось найти продажу с ID %d: %w", saleID, err)
	}

	// Обновляем поле Config
	sale.Config = newConfig
	if err := db.Save(&sale).Error; err != nil {
		return fmt.Errorf("ошибка обновления Config для продажи с ID %d: %w", saleID, err)
	}

	return nil
}

// -------------------------------- ТОВАРЫ --------------------------------
// добовление новый товар
func Addproduct(product Product) bool {
	result := db.Create(&product)
	if result.Error == nil {
		log.Println("Добавлен товар ", product.ID)
		return true
	} else {
		log.Println("Ошибка добавление данных", result.Error)
		return false
	}
}
func CountProductsOnSale() (int64, error) {
	var count int64
	// Выполняем запрос, чтобы посчитать количество продуктов на акции
	if err := db.Model(&Product{}).Where("is_on_sale = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Получение всех товаров
func Getproducts() *[]Product {
	var products []Product
	err := db.Find(&products)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return &[]Product{}
	}
	return &products
}

// Получение всех товаров
func GetProductId(id int) (bool, *Product) {
	var product Product
	res := db.Where("id = ?", id).First(&product)
	if res.RowsAffected > 0 {
		return true, &product
	}
	return false, &product
}

func AddProductToUser(username string, productId string) (bool, *Sale, error) {
	var user User
	var product Product

	// Находим пользователя по UserName
	if res := db.Where("user_name = ?", username).First(&user); res.RowsAffected == 0 {
		return false, nil, fmt.Errorf("пользователь не найден")
	}

	// Находим продукт по ProductId
	if res := db.Where("id = ?", productId).First(&product); res.RowsAffected == 0 {
		return false, nil, fmt.Errorf("продукт не найден")
	}

	// Создаем запись о покупке
	sale := Sale{
		Userid:    user.Id,
		Productid: product.Id,
	}

	// Если продукт имеет ограничение по трафику, добавляем оставшийся трафик
	if product.IsTraffic {
		sale.RemainingTraffic = float32(product.Traffic) // добавляем оставшийся трафик
	}

	// Если продукт имеет срок действия, устанавливаем дату окончания действия
	if product.IsTerm {
		// Предположим, что срок действия продукта — это количество дней от текущей даты
		sale.ExpirationDate = time.Now().Add(time.Duration(product.Term) * 24 * time.Hour)
	}

	// Сохраняем запись о покупке
	if res := db.Create(&sale); res.Error != nil {
		return false, nil, res.Error
	}

	return true, &sale, nil
}

func GetUserStatistics(username string) (float32, time.Time, int, error) {
	var user User
	var sales []Sale

	// Находим пользователя по UserName
	if res := db.Where("user_name = ?", username).First(&user); res.RowsAffected == 0 {
		return 0, time.Time{}, 0, fmt.Errorf("пользователь не найден")
	}

	// Получаем все активированные тарифы пользователя
	if res := db.Where("userid = ?", user.Id).Find(&sales); res.Error != nil {
		return 0, time.Time{}, 0, res.Error
	}

	// Переменные для подсчета
	var totalTraffic float32
	var latestExpiration time.Time
	var activatedCount int

	// Перебираем все покупки пользователя
	for _, sale := range sales {
		// Добавляем оставшийся трафик
		totalTraffic += sale.RemainingTraffic

		// Находим самую позднюю дату истечения тарифа
		if sale.ExpirationDate.After(latestExpiration) {
			latestExpiration = sale.ExpirationDate
		}

		// Считаем количество активированных тарифов
		activatedCount++
	}

	// Возвращаем результаты
	return totalTraffic, latestExpiration, activatedCount, nil
}

// изменение названия товара
func UpdProductName(id uint, name string) bool {
	var product = Product{Id: id}
	err := db.First(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	product.Name = name
	err = db.Save(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}

// изменение цены товара
func UpdProductPrice(id, price uint) bool {
	var product = Product{Id: id}
	err := db.First(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	product.NowPrice = price
	err = db.Save(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}

// изменение времени товара
func UpdProductTerm(id, term uint) bool {
	var product = Product{Id: id}
	err := db.First(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	product.Term = term
	err = db.Save(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}

// Удаление товара
func Dellproducts(id uint) bool {
	err := db.Delete(&Product{}, id)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}
