package main

import (
	"Project/utils"
	"fmt"
	"log"
	"os"
)

func main() {
	wg_client := utils.WireGuardConfig{}
	id := 2
	//начальное сообщение
	firstMess := "Доступные команды:\n" +
		"1 Запуск wireguard\n" +
		"2 Добавление клиента\n" +
		"3 Остановка клиента\n" +
		"4 Активация клиента\n" +
		"5 Удаление клиента\n" +
		"6 Все клиенты\n" +
		"0 В меню\n" +
		"-1 Выйти\n" +
		"-2 Удалить и отключить wiregurad,выйти"
	//прием команд
	var comand int
	fmt.Printf(firstMess)
br:
	for true {
		fmt.Fscan(os.Stdin, &comand)
		switch comand {
		case 1: // задание основных настроек
			port := ""
			fmt.Println("Введите порт,или 0 для порта 51820: ")
			fmt.Scanln(&port)
			if port == "0" {
				wg_client.ListenPort = "51820"
			} else {
				wg_client.ListenPort = port
			}
			ip := ""
			fmt.Println("Введите ip сервера : ")
			fmt.Scanln(&ip)
			wg_client.Endpoint = ip + ":" + wg_client.ListenPort
			inter := ""
			fmt.Println("Введите имя интерфейса,или 0 для eth0 : ")
			fmt.Scanln(&inter)
			if inter == "0" {
				wg_client.InterName = "eth0"
			} else {
				wg_client.InterName = inter
			}
			fmt.Println("Введите токен бота : ")
			fmt.Scanln(&wg_client.BotToken)
			fmt.Println("Параметры успешно заданы")
			wg_client.CollectTraffic()
			wg_client.GenServerKeys()
			fmt.Println()
			fmt.Println("Созданный приватный ключ : ", wg_client.PrivateKey)
			fmt.Println("Созданный публичный ключ : ", wg_client.PublicKey)
			fmt.Println()
			fmt.Println("Создана конфигурация с такими параметрами:")
			fmt.Println("Порт : ", wg_client.ListenPort)
			fmt.Println("Приватный ключ сервера : ", wg_client.PrivateKey)
			fmt.Println("Имя интерфейса : ", wg_client.InterName)
			fmt.Println("Адресом : 10.0.0.1/24")
			fmt.Println("Эндпоинт : ", wg_client.Endpoint)
			fmt.Println()
			wg_client.GenerateWireGuardConfig()
			wg_client.WireguardStart()
			log.Printf("Соединение wireguard запущено")
		case 2: //Добавление клиента
			client, _ := wg_client.AddWireguardClient(id)
			fmt.Println("Данные клиента : ")
			fmt.Println("Адресс : ", client.AddressClient)
			fmt.Println("Публичный ключ : ", client.PublicClientKey)
			fmt.Println("Приватный ключ : ", client.PrivateClientKey)
			tgID := -1
			fmt.Println("Введите тг id клиента или -1, если отправлять конфиг не нужно: ")
			fmt.Scanln(&tgID)
			if tgID != -1 {
				client.TgId = tgID
				wg_client.Clients[id] = client
				wg_client.SendConfigToUserTg(id)
				fmt.Println("Данные отправлены")
			}
			id++
		case 3: //Остановка клиента
			usID := 0
			fmt.Println("Введите id клиента для остановки: ")
			fmt.Scanln(&usID)
			wg_client.StopClient(usID)
			fmt.Println("Клиента остановлен")
		case 4: //Активация клиента
			usID := 0
			fmt.Println("Введите id клиента для активации: ")
			fmt.Scanln(&usID)
			wg_client.ActClient(usID)
			fmt.Println("Клиента активирован")
		case 5:
			usID := 0
			fmt.Println("Введите id клиента для удаления: ")
			fmt.Scanln(&usID)
			wg_client.DeleteClient(usID)
			fmt.Println("Клиента остановлен")
		case 6: //все клиенты
			cleints := wg_client.AllClients()
			fmt.Println(cleints)
		case 7: //Отправить данные клиента
			usID := -1
			fmt.Println("Введите id клиента или -1, если отправлять конфиг не нужно: ")
			fmt.Scanln(&usID)
			if usID != -1 {
				fmt.Println(usID)
				wg_client.SendConfigToUserTg(usID)
				fmt.Println("Данные отправлены")
			}
		case -2:
			wg_client.DropWireguard()
			fmt.Println("Конфигурации и скрипт удален")
			break br
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
