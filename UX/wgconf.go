package UX

import (
	"Project/db"
	"Project/wireguard_go_ubuntu"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

// настройка wireguard
var Filename = "data.json"
var Wg_client = wireguard_go_ubuntu.WireGuardConfig{}

func init() {
	if _, err := os.Stat(Filename); os.IsNotExist(err) {
		log.Printf("File %s does not exist. Executing Autostart.\n", Filename)
		// Выполняем Autostart
		Wg_client.Autostart()
		Wg_client.SaveToFile(Filename)
	} else {
		Wg_client.LoadFromFile(Filename)

	}
}

// Добавление клиента, возвращает peerStr и config
func AddConf(confid int) (string, string, error) {
	// Добавляем клиента через Wg_client
	client, _ := Wg_client.AddWireguardClient(confid)

	// Сохраняем изменения в файл
	Wg_client.SaveToFile(Filename)

	// Логируем успех
	log.Printf("Клиент с ID %d добавлен. Конфигурация: %s", confid, client.Config)

	// Возвращаем peerStr и config
	return client.PeerStr, client.Config, nil
}

// удаление клиента
func DeleteConf(confid int) {
	Wg_client.DeleteClient(confid)
	log.Printf("Клиент с ID %d удален", confid)
	Wg_client.SaveToFile(Filename)
}

// остановка клиента
func StopConf(confid int) {
	Wg_client.StopClient(confid)
	log.Printf("Клиент с ID %d остановлен", confid)
	Wg_client.SaveToFile(Filename)
}

// GetClientTraffic возвращает трафик для конкретного клиента по его id в гигабайтах
func GetConfTraffic(confid string) (int, error) {
	// Получаем все данные о трафике
	trafficData, err := Wg_client.CollectTraffic()
	if err != nil {
		return 0, err
	}

	// Проверяем, есть ли трафик для клиента с данным id
	peerTraffic, exists := trafficData[confid]
	if !exists {
		return 0, fmt.Errorf("client with id %s not found", confid)
	}

	// Конвертируем трафик в гигабайты и возвращаем
	rxGB := peerTraffic.TrafficRx / (1024 * 1024 * 1024) // В гигабайтах
	txGB := peerTraffic.TrafficTx / (1024 * 1024 * 1024) // В гигабайтах

	// Возвращаем общий трафик (полученный + отправленный)
	totalTrafficGB := int(rxGB + txGB)
	return totalTrafficGB, nil
}

// GetTraffic возвращает трафик всех клиентов в виде map[id]трафик в гигабайтах
func GetTraffic() (map[int]int, error) {
	// Получаем все данные о трафике
	trafficData, err := Wg_client.CollectTraffic()
	if err != nil {
		return nil, err
	}

	// Создаем карту для возврата
	clientTraffic := make(map[int]int)

	// Проходим по всем данным о трафике
	for clientID, peerTraffic := range trafficData {
		rxGB := peerTraffic.TrafficRx / (1024 * 1024 * 1024) // В гигабайтах
		txGB := peerTraffic.TrafficTx / (1024 * 1024 * 1024) // В гигабайтах

		// Сохраняем общий трафик для клиента (полученный + отправленный)
		clientTrafficID, err := strconv.Atoi(clientID)
		if err != nil {
			log.Printf("Error converting client id %s to int: %v", clientID, err)
			continue
		}

		totalTrafficGB := int(rxGB + txGB)
		clientTraffic[clientTrafficID] = totalTrafficGB
	}

	return clientTraffic, nil
}

func SendConf(confid string, tgid int64) {
	txt := fmt.Sprintf("```%s```", confid)
	msg := tgbotapi.NewMessage(tgid, txt)
	TelegramBot.Send(msg)
}

func UpdateTraffic() {
	// Получаем данные о трафике
	trafficData, err := GetTraffic()
	if err != nil {
		log.Printf("Failed to get traffic data: %v", err)
		return
	}

	// Проходим по всем данным о трафике и обновляем базу данных
	for clientID, usedTraffic := range trafficData {
		var sale db.Sale
		if err := db.DB.First(&sale, "id = ?", clientID).Error; err != nil {
			log.Printf("Failed to find sale with ID %d: %v", clientID, err)
			continue
		}

		// Вычитаем использованный трафик за час из оставшегося
		remainingTraffic := sale.RemainingTraffic - float32(usedTraffic)
		if remainingTraffic < 0 {
			remainingTraffic = 0
		}

		sale.RemainingTraffic = remainingTraffic

		// Сохраняем обновленные данные в базе
		if err := db.DB.Save(&sale).Error; err != nil {
			log.Printf("Failed to update traffic for sale ID %d: %v", clientID, err)
		}
	}
}
