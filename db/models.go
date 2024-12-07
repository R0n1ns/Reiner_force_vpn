package db

import (
	"gorm.io/gorm"
	"time"
)

// Модель для пользователей
type User struct {
	gorm.Model
	Id        uint   `gorm:"primaryKey;autoIncrement;not null;unique"` // ид системы
	Tgid      uint   `gorm:"not null;unique"`                          // ид тг для входа
	Mail      string `gorm:"default:'null'"`
	Password  string `gorm:"default:'null'"` // пароль
	UserName  string `gorm:"not null;unique"`
	Isblocked bool   `gorm:"default:false"` // заблокирован ли пользователь
	IsAdmin   bool   `gorm:"default:false"` // админ ли пользователь
	Purchases []Sale `gorm:"foreignKey:Userid"`
}

// Модель для продуктов
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

// Модель для покупок
type Sale struct {
	gorm.Model
	Id                  uint      `gorm:"primaryKey;autoIncrement;not null;unique"`
	Userid              uint      `gorm:"not null"` // внешний ключ для пользователя
	Peer                string    `gorm:"not null"`
	Config              string    `gorm:"not null"`
	User                User      `gorm:"foreignKey:Userid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Productid           uint      `gorm:"not null"` // внешний ключ для продукта
	Product             Product   `gorm:"foreignKey:Productid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ISFrozen            bool      `gorm:"default:false"`
	ExpirationFrozeDate uint      `gorm:""`
	ExpirationDate      time.Time `gorm:""` // колонка для даты истечения тарифа
	RemainingTraffic    float32   `gorm:""` // колонка оставшегося трафика
}
