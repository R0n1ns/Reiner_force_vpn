package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// User модель для пользователей
type User struct {
	gorm.Model
	Id        uint   `gorm:"primaryKey;autoIncrement;not null;unique"`
	Tgid      uint   `gorm:"not null;unique"`
	Mail      string `gorm:"default:null"`
	Password  string `gorm:"default:null"`
	UserName  string `gorm:"not null;unique"`
	Isblocked bool   `gorm:"default:false"`
	IsAdmin   bool   `gorm:"default:false"`
	Purchases []Sale `gorm:"foreignKey:Userid"`
}

// Product модель для продуктов
type Product struct {
	gorm.Model
	Id        uint   `gorm:"primaryKey;autoIncrement;not null"`
	Name      string `gorm:"not null"`
	NowPrice  uint   `gorm:"not null"`
	LastPrice uint   `gorm:"default:0"`
	IsOnSale  bool   `gorm:"default:false"`
	IsTerm    bool   `gorm:"default:false"`
	Term      uint   `gorm:"default:0"`
	IsTraffic bool   `gorm:"default:false"`
	Traffic   uint   `gorm:"default:0"`
}

// Sale модель для покупок
type Sale struct {
	gorm.Model
	Id                  uint      `gorm:"primaryKey;autoIncrement;not null;unique"`
	Userid              uint      `gorm:"not null"`
	Peer                string    `gorm:"not null"`
	Config              string    `gorm:"not null"`
	User                User      `gorm:"foreignKey:Userid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Productid           uint      `gorm:"not null"`
	Product             Product   `gorm:"foreignKey:Productid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ISFrozen            bool      `gorm:"default:false"`
	ExpirationFrozeDate uint      `gorm:"default:0"`
	ExpirationDate      time.Time `gorm:""`
	RemainingTraffic    float32   `gorm:"default:0"`
}

// Log модель для логов
type Log struct {
	gorm.Model
	LogName    string    `gorm:"not null"`
	LogType    string    `gorm:"not null"`
	LoggedTime time.Time `gorm:"autoCreateTime:nano"`
}

// main функция для подключения и миграции базы данных
func main() {
	// Настройки подключения
	dsn := "host=localhost user=vpn password=vpn dbname=vpn port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	// Подключение к PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Включение имен таблиц во множественном числе
		},
	})

	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	log.Println("Подключение к базе данных успешно.")

	// Авто-миграция для создания таблиц
	err = db.AutoMigrate(&User{}, &Product{}, &Sale{}, &Log{})
	if err != nil {
		log.Fatal("Ошибка миграции:", err)
	}

	log.Println("Миграция успешно завершена.")
}
