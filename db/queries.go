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

// миграция баз данных
func Migrations() {
	//снос таблиц
	//db.Migrator().DropTable(&User{})
	//db.Migrator().DropTable(&Products{})
	//db.Migrator().DropTable(&Config{})

	//выполнение миграций
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Products{})
	db.AutoMigrate(&Config{})

	//отчет
	log.Println("Миграции выполнены")
}

// -------------------------------- ПОЛЬЗОВАТЕЛИ --------------------------------

// добовление нового пользователя
func Adduser(user User) bool {
	result := db.Create(&user) // pass pointer of data to Create
	if result.Error == nil {
		log.Println("Добавлен пользователь ", user.ID)
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

// получение пользователя
func GetUserUsername(usrname string) (bool, User) {
	var user User
	res := db.Find(&user, "UserName = ?", usrname)
	if res.Error == nil {
		return true, user
	} else {
		return false, User{}
	}
}

//пполучение данных пользователя
//изменения пользователя
//изменения пароля пользователя

// -------------------------------- ТОВАРЫ --------------------------------
// добовление новый товар
func Addproduct(product Products) bool {
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
func Getproducts() *[]Products {
	var products []Products
	err := db.Find(&products)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return &[]Products{}
	}
	return &products
}

// изменение названия товара
func UpdProductName(id uint, name string) bool {
	var product = Products{Id: id}
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
	var product = Products{Id: id}
	err := db.First(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	product.Price = price
	err = db.Save(&product)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}

// изменение времени товара
func UpdProductTerm(id, term uint) bool {
	var product = Products{Id: id}
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
	err := db.Delete(&Products{}, id)
	if err.Error != nil {
		log.Println("Ошибка получения продуктов", err)
		return false
	}
	return true
}
