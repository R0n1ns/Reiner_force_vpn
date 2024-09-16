package micro

import (
	"Project/db"
	"fmt"
	"os"
)

func Main_micro() {
	//начальное сообщение
	firstMess := "Доступные команды:\n" +
		"1 Все пользователи\n" +
		"2 Добавление нового пользователя\n" +
		"3 Удалить пользователя по id" +
		"4 Все товары\n" +
		"5 Добавление нового товара\n" +
		"6 Удалить товар по id\n" +
		"0 В меню\n" +
		"-1 Выключить\n"
	//прием команд
	var comand int
	fmt.Printf(firstMess)
br:
	for true {
		fmt.Fscan(os.Stdin, &comand)
		switch comand {
		case 1:
			//получение всех пользователей
			users := db.GetUsers()
			fmt.Println("Вот все пользователи:\n")
			for _, user := range *users {
				fmt.Println("ID пользователя :", user.Id, " Почта пользователя :", user.Mail, " Телеграмм id пользователя :", user.Tgid)
			}
		case 2:
			fmt.Println(2)
		case 0:
			fmt.Println(firstMess)
		case -1:
			fmt.Println("Скрипт остановлен")
			break br
		default:
			fmt.Println("Команда не найдена\n" + firstMess)

		}
	}

}
