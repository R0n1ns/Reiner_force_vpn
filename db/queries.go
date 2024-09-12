package db

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var load_err = godotenv.Load("config.env")
var db_path = os.Getenv("DB_PATH")

var db, err = gorm.Open(postgres.Open(db_path), &gorm.Config{})

// миграция баз данных
func Migrations() {
	if err != nil {
		log.Println("Ошибка подключения к базе", err)
	} else {
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
}

// добовление нового пользователя
func Adduser(user User) int {

	result := db.Create(&user) // pass pointer of data to Create

	if result.Error == nil {
		log.Println("Добавлен пользователь ", user.ID)
		return int(user.ID)
	} else {
		log.Println("Ошибка добавдение данных", result.Error)
		return -1
	}
}

// получить всех пользователей
func GetUsers() *[]User {
	// Get all record
	var users []User
	db.Find(&users)
	return &users
}

// добовление новый товар
func Addproduct(price uint, term uint, name string) int {
	product := Products{Price: price, Term: term, Name: name}
	result := db.Create(&product)
	if result.Error == nil {
		log.Println("Добавлен товар ", product.ID)
		return int(product.ID)
	} else {
		log.Println("Ошибка добавление данных", result.Error)
		return -1
	}
}

func Getproducts() *[]Products {
	// Get all record
	var products []Products
	result := db.Find(&products)
	//запись в удобную форму
	get_productns := make([]Products, result.RowsAffected)
	//обработка ответа
	if result.Error != nil {
		log.Println(result.Error)
	} else {
		//перебор и добавление продуктов в виде id:цена
		for i, product := range products {
			get_productns[i] = Products{Name: product.Name, Price: product.Price, Term: product.Term}
		}
	}
	return &get_productns

	//для теста
	//products := db.Getproducts()
	//for _, product := range products {
	//	log.Println("Название ", product.Name, "Цена ", product.Price, "Время ", product.Term)
	//}
}
