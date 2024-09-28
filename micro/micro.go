package main

import (
	"Project/utils"
	"fmt"
	"log"
	"os"
)

func main() {
	wg_client := utils.WireGuardConfig{}
	//начальное сообщение
	firstMess := "Доступные команды:\n" +
		"1 Запуск wireguard\n" +
		"2 Добавление клиента\n" +
		"3 Остановка клиента\n" +
		"4 Удаление клиента\n" +
		"5 Все клиенты\n" +
		"6 Отправить конфигурацию клиента\n" +
		"0 В меню\n" +
		"-1 Выключить\n" +
		"-2 Удалить и отключить wiregurad,выйти"
	//прием команд
	var comand int
	fmt.Printf(firstMess)
br:
	for true {
		fmt.Fscan(os.Stdin, &comand)
		switch comand {
		case 1: // задание основных настроек
			fmt.Println("Введите порт: ")
			fmt.Scanln(&wg_client.ListenPort)
			ip := ""
			fmt.Println("Введите ip : ")
			fmt.Scanln(&ip)
			wg_client.Endpoint = ip + ":" + wg_client.ListenPort
			fmt.Println("Введите имя интерфейса : ")
			fmt.Scanln(&wg_client.InterName)
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
		case 2:
		case 3:
		case 4:
		case 5:
			clientConf := utils.WireGuardClientConfig{
				ServerPubKey: Publickkey,
				Endpoint:     (Serv_ip + ":" + Port),
			}
			var client utils.Client
			lastUserConfig, client = utils.AddWireguardClient(clientConf, LastClientAdr)
			Clients.Clients[LastClientAdr] = &client
			LastClientAdr++
		case 6: //
			utils.SendConfigToUserTg(lastUserConfig, ChatId, Token)
			fmt.Println("Конфиг отправлен")
		case -2:
			utils.DropWireguard()
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
