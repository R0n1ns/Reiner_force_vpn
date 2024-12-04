package db

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
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
	//db.Migrator().DropTable(&Config{})

	//выполнение миграций
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Product{})
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
			"PlanName":         sale.Product.Name,
			"Status":           "Активен", // Логика проверки статуса
			"RemainingTraffic": sale.RemainingTraffic,
			"ExpirationDate":   sale.ExpirationDate.Format("02.01.2006"),
			"IsRenewable":      !sale.ISFrozen,
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

//пполучение данных пользователя
//изменения пользователя
//изменения пароля пользователя

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
