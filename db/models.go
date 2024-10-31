package db

import "gorm.io/gorm"

// Модель для пользователей
type User struct {
	gorm.Model
	Id        uint     `gorm:"primaryKey;autoIncrement;not null;unique"` //ид системы
	Tgid      uint     `gorm:"primaryKey;not null;unique"`               //ид тг для входа
	Mail      string   `gorm:"default:'null'"`
	Password  string   `gorm:"default:'null'"` //пароль
	UserName  string   `gorm:"not null;unique"`
	Isblocked bool     `gorm:"default:false"` //заблокирован ли пользователь
	Config    []Config `gorm:"foreignKey:Userid"`
}

type Products struct {
	gorm.Model
	Id    uint `gorm:"primaryKey;autoIncrement;not null"`
	Name  string
	Price uint `gorm:"not null"`
	Term  uint `gorm:"default:1;not null"`
}

type Config struct {
	gorm.Model
	Id     uint `gorm:"primaryKey;autoIncrement;not null;unique"`
	Userid uint `gorm:"not null"`
}
