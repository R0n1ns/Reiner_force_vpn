package main

import (
	"Project/wireguard_go_ubuntu"
	"fmt"
	"log"
)

var filename = "data.json"
var Wg_client = wireguard_go_ubuntu.WireGuardConfig{}

func init() {
	Wg_client.LoadFromFile(filename)
}

// Функция для запроса целого числа от пользователя
func inputInt(prompt string) int {
	var value int
	fmt.Print(prompt)
	fmt.Scanln(&value)
	return value
}

// Функция для запроса строки от пользователя
func inputString(prompt string) string {
	var value string
	fmt.Print(prompt)
	fmt.Scanln(&value)
	return value
}

func handleStartWireguard(Wg_client *wireguard_go_ubuntu.WireGuardConfig) {
	tp := inputInt("если хотите автоматическую настройку введите 0, ручной режим 1: ")
	if tp != 0 {
		Wg_client.ListenPort = inputString("Введите порт, или 0 для порта 51820: ")
		if Wg_client.ListenPort == "0" {
			Wg_client.ListenPort = "51820"
		}
		Wg_client.Endpoint = inputString("Введите ip сервера: ") + ":" + Wg_client.ListenPort
		Wg_client.InterName = inputString("Введите имя интерфейса, или 0 для eth0: ")
		if Wg_client.InterName == "0" {
			Wg_client.InterName = "eth0"
		}
		Wg_client.BotToken = inputString("Введите токен бота: ")

		fmt.Println("Параметры успешно заданы")
		Wg_client.CollectTraffic()
		Wg_client.GenServerKeys()
		fmt.Printf("Созданный приватный ключ: %s\nСозданный публичный ключ: %s\n", Wg_client.PrivateKey, Wg_client.PublicKey)

		Wg_client.GenerateWireGuardConfig()
		Wg_client.WireguardStart()
		log.Printf("Соединение wireguard запущено")
	} else {
		Wg_client.Autostart()
		Wg_client.BotToken = inputString("Введите токен бота: ")
		fmt.Printf("Созданный приватный ключ: %s\nСозданный публичный ключ: %s\n", Wg_client.PrivateKey, Wg_client.PublicKey)
	}
}

func handleAddClient(Wg_client *wireguard_go_ubuntu.WireGuardConfig, id *int) {
	client, _ := Wg_client.AddWireguardClient(*id)
	fmt.Printf("Данные клиента:\nАдрес: %s\nПубличный ключ: %s\nПриватный ключ: %s\n",
		client.AddressClient, client.PublicClientKey, client.PrivateClientKey)

	tgID := inputInt("Введите тг id клиента или -1, если отправлять конфиг не нужно: ")
	if tgID != -1 {
		client.TgId = tgID
		Wg_client.Clients[*id] = client
		Wg_client.SendConfigToUserTg(*id)
		fmt.Println("Данные отправлены")
	}
	(*id)++
}

func handleClientAction(Wg_client *wireguard_go_ubuntu.WireGuardConfig, action func(int), prompt string) {
	usID := inputInt(prompt)
	action(usID)
}

func printAllClients(Wg_client *wireguard_go_ubuntu.WireGuardConfig) {
	clients := Wg_client.AllClients()
	fmt.Println(clients)
}

func handleSendClientData(Wg_client *wireguard_go_ubuntu.WireGuardConfig) {
	usID := inputInt("Введите id клиента или -1, если отправлять конфиг не нужно: ")
	if usID != -1 {
		Wg_client.SendConfigToUserTg(usID)
		fmt.Println("Данные отправлены")
	}
}

func handleDropWireguard(Wg_client *wireguard_go_ubuntu.WireGuardConfig) {
	Wg_client.DropWireguard()
	fmt.Println("Конфигурации и скрипт удален")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// Вывод ошибки
			fmt.Println("Произошла ошибка:", r)
			// Вызов функции снова
			main()
		}
	}()

	// Отложенное сохранение данных перед выходом
	defer func() {
		err := Wg_client.SaveToFile(filename)
		if err != nil {
			log.Printf("Ошибка при сохранении файла: %v", err)
		} else {
			fmt.Println("Конфигурация успешно сохранена.")
		}
	}()

	id := 2
	firstMess := `Доступные команды:
1. Запуск wireguard
2. Добавление клиента
3. Остановка клиента
4. Активация клиента
5. Удаление клиента
6. Все клиенты
7. Отправить данные клиента
0. В меню
-1. Выйти
-2. Удалить и отключить wireguard, выйти`

	fmt.Println(firstMess)

	commands := map[int]func(){
		1: func() { handleStartWireguard(&Wg_client) },
		2: func() { handleAddClient(&Wg_client, &id) },
		3: func() {
			handleClientAction(&Wg_client, Wg_client.StopClient, "Введите id клиента для остановки: ")
		},
		4: func() {
			handleClientAction(&Wg_client, Wg_client.ActClient, "Введите id клиента для активации: ")
		},
		5: func() {
			handleClientAction(&Wg_client, Wg_client.DeleteClient, "Введите id клиента для удаления: ")
		},
		6:  func() { printAllClients(&Wg_client) },
		7:  func() { handleSendClientData(&Wg_client) },
		-2: func() { handleDropWireguard(&Wg_client) },
		0:  func() { fmt.Println(firstMess) },
	}

	for {
		comand := inputInt("Введите команду: ")
		if comand == -1 {
			fmt.Println("Скрипт остановлен")
			// Прерываем цикл, вместо вызова os.Exit
			break
		}

		if action, exists := commands[comand]; exists {
			action()
		} else {
			fmt.Println("Команда не найдена\n" + firstMess)
		}
	}
}
