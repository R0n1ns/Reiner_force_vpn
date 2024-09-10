package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var load_err = godotenv.Load("config.env")
var db_path = os.Getenv("DB_PATH")

var db, err = gorm.Open(postgres.Open(db_path), &gorm.Config{})

// миграция баз данных
func Migrations() {
	if err != nil {
		fmt.Println("Ошибка подключения к базе", err)
	} else {
		//снос таблиц
		db.Migrator().DropTable(&User{})
		db.Migrator().DropTable(&Products{})
		db.Migrator().DropTable(&Config{})

		//выполнение миграций
		db.AutoMigrate(&User{})
		db.AutoMigrate(&Products{})
		db.AutoMigrate(&Config{})

		//отчет
		fmt.Println("Миграции выполнены")
	}
}

// добовление нового пользователя
func Adduser(tgid uint, password string) int {
	user := User{Tgid: tgid, Password: password}

	result := db.Create(&user) // pass pointer of data to Create

	if result.Error == nil {
		fmt.Println("Добавлен пользователь ", user.ID)
		return int(user.ID)
	} else {
		fmt.Println("Ошибка добавдение данных", result.Error)
		return -1
	}
}

// добовление новый товар
func Addproduct(price uint, term uint, name string) int {
	product := Products{Price: price, Term: term, Name: name}

	result := db.Create(&product)

	if result.Error == nil {
		fmt.Println("Добавлен товар ", product.ID)
		return int(product.ID)
	} else {
		fmt.Println("Ошибка добавление данных", result.Error)
		return -1
	}
}

func Getproducts() []Products {
	// Get all record
	var products []Products
	result := db.Find(&products)
	//запись в удобную форму
	get_productns := []Products{}
	//обработка ответа
	if result.Error != nil {
		fmt.Println(result.Error)
	} else {
		//перебор и добавление продуктов в виде id:цена
		for _, product := range products {
			get_productns = append(get_productns, Products{Name: product.Name, Price: product.Price, Term: product.Term})
		}
	}
	return get_productns

	//для теста
	//products := db.Getproducts()
	//for _, product := range products {
	//	fmt.Println("Название ", product.Name, "Цена ", product.Price, "Время ", product.Term)
	//}
}
