package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/telebot.v3"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Структура для конфигурации пира
type PeerConfig struct {
	PublicKey  string
	AllowedIPs string
	Endpoint   string
}

// Структура для клиента
type Client struct {
	Id               int
	Status           bool
	AddressClient    string
	PubkeyPath       string
	PrivkeyPath      string
	PrivateClientKey string
	PublicClientKey  string
	Peer             PeerConfig
	PeerStr          string
	Config           string
	TgId             int
}

// Управление сервером WireGuard
type WireGuardConfig struct {
	PrivateKey string
	PublicKey  string
	Endpoint   string
	ListenPort string
	InterName  string
	BotToken   string
	Clients    map[int]Client // Используем указатели на клиентов
}

// ------------------------ методы для клиентов ------------------------
// Остановка клиента
func (wg WireGuardConfig) StopClient(id int) {
	client, exists := wg.Clients[id]
	if !exists {
		log.Printf("Клиент с id %d не найден", id)
		return
	}
	defer func() { wg.Clients[id] = client }()
	//fmt.Println(wg.Clients[id])

	filePath := "/etc/wireguard/wg0.conf"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Ошибка чтения файла конфигурации: %v", err)
		return
	}
	defer restWireguard()

	fileContent := string(content)
	client.Status = false
	updatedContent := strings.Replace(fileContent, client.PeerStr, "", 1)

	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Printf("Ошибка записи файла конфигурации: %v", err)
		return
	}

	log.Printf("Клиент с id %d остановлен", id)
}

// Активация клиента
func (wg WireGuardConfig) ActClient(id int) {
	client, exists := wg.Clients[id]
	if !exists {
		log.Printf("Клиент с id %d не найден", id)
		return
	}
	//fmt.Println(wg.Clients[id])
	defer restWireguard()

	filePath := "/etc/wireguard/wg0.conf"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Ошибка чтения файла конфигурации: %v", err)
		return
	}
	defer func() { wg.Clients[id] = client }()

	client.Status = true
	updatedContent := string(content) + "\n" + client.PeerStr

	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Printf("Ошибка записи файла конфигурации: %v", err)
		return
	}

	log.Printf("Клиент с id %d активирован", id)
}

// Удаление клиента
func (wg *WireGuardConfig) DeleteClient(id int) {
	wg.StopClient(id)

	//err := os.Remove(fmt.Sprintf("/etc/wireguard/wg_client_%d_private", id))
	//if err != nil {
	//	log.Printf("Не удалось удалить файл: %v", err)
	//}
	//
	//err = os.Remove(fmt.Sprintf("/etc/wireguard/wg_client_%d_public", id))
	//if err != nil {
	//	log.Printf("Не удалось удалить файл: %v", err)
	//}

	delete(wg.Clients, id)
}

// вывод всех клиентов
func (clients *WireGuardConfig) AllClients() string {
	text := ""
	for id, client := range clients.Clients {
		var stat string
		if client.Status {
			stat = "Активен"
		} else {
			stat = "Остановлен"
		}
		text += fmt.Sprintf("Клиент %d статус %s адресс %s \n", id, stat, client.AddressClient)
	}
	return text

}
func (wg *WireGuardConfig) Autostart() {
	wg.RandomPort()
	wg.GetIPAndInterfaceName()
	wg.GenServerKeys()
	wg.GenerateWireGuardConfig()
	// wg_client.CollectTraffic()
	wg.WireguardStart()
}

// генерируем ключи
func (wg *WireGuardConfig) GenServerKeys() {
	//генерируем ключи
	var privateKey bytes.Buffer
	cmd := exec.Command("wg", "genkey")
	cmd.Stdout = &privateKey
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}
	// Сохраняем приватный ключ в переменную
	privatekey := strings.ReplaceAll(privateKey.String(), "\n", "")
	// Используем приватный ключ для генерации публичного ключа
	var publicKey bytes.Buffer
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = &privateKey
	cmd.Stdout = &publicKey

	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to generate public key: %v", err)
	}
	publickkey := strings.ReplaceAll(publicKey.String(), "\n", "")
	//запись
	os.WriteFile("/etc/wireguard/privatekey", []byte(privatekey), 0600)
	os.WriteFile("/etc/wireguard/publickey", []byte(publickkey), 0600)
	// Сохраняем публичный ключ в переменную
	time.Sleep(time.Second * 5)
	wg.PublicKey = publickkey
	wg.PrivateKey = privatekey
}
func (wg *WireGuardConfig) RandomPort() {
	wg.ListenPort = strconv.Itoa(rand.Intn(100000))
	//fmt.Println(wg.ListenPort)
}

type NetworkInterface struct {
	Name string
	IsUp bool
	IPs  []string
}

func (cfg *WireGuardConfig) GetIPAndInterfaceName() error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы или интерфейсы без нужных флагов
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Пропускаем не IPv4 адреса
			if ip == nil || ip.To4() == nil {
				continue
			}

			cfg.InterName = iface.Name
			cfg.Endpoint = ip.String() + ":" + cfg.ListenPort
			return nil
		}
	}

	return fmt.Errorf("не удалось найти подходящий IP-адрес и интерфейс")
}

// Функция для определения, является ли интерфейс проводным
func isWiredInterface(name string) bool {
	return strings.HasPrefix(name, "e") || strings.Contains(name, "eth") || strings.Contains(name, "en")
}

// Функция для определения, является ли интерфейс беспроводным
func isWirelessInterface(name string) bool {
	return strings.HasPrefix(name, "w") || strings.Contains(name, "wl") || strings.Contains(name, "wlan")
}

// Генерация конфигурации WireGuard
func (wg *WireGuardConfig) GenerateWireGuardConfig() {
	//генерация конфига для мурвера
	tmpl := `[Interface]
PrivateKey = {{.PrivateKey}}
Address = 10.0.0.1/24
ListenPort = {{.ListenPort}}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{.InterName}} -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{.InterName}} -j MASQUERADE`

	t := template.Must(template.New("wgConfig").Parse(tmpl))

	var buf bytes.Buffer
	if err := t.Execute(&buf, wg); err != nil {
	}
	// Генерация случайного числа от 0 до 99999
	// Открытие файла для записи
	filePath := "/etc/wireguard/wg0.conf"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	// Запись данных в файл
	_, err = file.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("Ошибка записи в файл: %v\n", err)
		return
	}

	//log.Println("Конфигурация wireguard сгенерирована")

}

func (wg *WireGuardConfig) WireguardStart() {
	port := wg.ListenPort
	// настройка форвардинг
	file, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open /etc/sysctl.conf: %v", err)
	}
	defer file.Close()

	// Записываем строку в файл
	_, err = file.WriteString("net.ipv4.ip_forward=1\n")
	if err != nil {
		log.Fatalf("failed to write to /etc/sysctl.conf: %v", err)
	}
	prt := fmt.Sprintf("%s/udp", port)
	// Выполняем команду `sysctl -p` для применения изменений
	cmd := exec.Command("ufw", "allow", prt)
	err = cmd.Run()
	if err != nil {
		log.Printf("failed to apply sysctl changes: %v", err.Error())
	}
	// Выполняем команду `sysctl -p` для применения изменений
	cmd = exec.Command("sysctl", "-p")
	err = cmd.Run()
	if err != nil {
		log.Printf("failed to apply sysctl changes: %v", err.Error())
	}
	//включсение wireguard
	cmd = exec.Command("systemctl", "enable", "wg-quick@wg0.service")
	cmd.Run()
	err = cmd.Err
	if err != nil {
		log.Printf("failed to create keys : %v", err.Error())
	}
	//старт wireguard
	cmd = exec.Command("systemctl", "start", "wg-quick@wg0.service")
	cmd.Run()
	err = cmd.Err
	if err != nil {
		log.Printf("failed to create keys : %v", err.Error())
	}
	//log.Printf("Соединение wireguard запущено")
}
func restWireguard() {
	cmd := exec.Command("systemctl", "restart", "wg-quick@wg0")
	err := cmd.Run()
	if err != nil {
		log.Printf("failed to apply sysctl changes: %v", err.Error())
	}

}

// Добавление клиента WireGuard
func (wg *WireGuardConfig) AddWireguardClient(clientID int) (Client, int) {
	// Инициализация карты клиентов, если она nil
	if wg.Clients == nil {
		wg.Clients = make(map[int]Client)
	}
	defer restWireguard()
	// Проверяем, существует ли клиент
	client, exists := wg.Clients[clientID]
	if !exists {
		client = Client{Id: clientID}
		wg.Clients[clientID] = client
	}
	defer func() { wg.Clients[clientID] = client }()
	// Генерация ключей для клиента
	var privateKey, publicKey bytes.Buffer
	cmd := exec.Command("wg", "genkey")
	cmd.Stdout = &privateKey
	err := cmd.Run()

	client.PrivateClientKey = strings.TrimSpace(privateKey.String())
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = &privateKey
	cmd.Stdout = &publicKey
	err = cmd.Run()

	client.PublicClientKey = strings.TrimSpace(publicKey.String())
	client.AddressClient = fmt.Sprintf("10.0.0.%d/24", clientID)
	client.Peer.Endpoint = wg.Endpoint
	client.Peer.PublicKey = wg.PublicKey
	peer := fmt.Sprintf("\n[Peer]\nPublicKey = %s\nAllowedIPs = %s\n", strings.TrimSpace(publicKey.String()), fmt.Sprintf("10.0.0.%d/24", clientID))
	client.PeerStr = peer
	filePath := "/etc/wireguard/wg0.conf"
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(peer); err != nil {
		panic(err)
	}
	client.Status = true
	// Генерация и сохранение конфигурации клиента
	clientConfig := fmt.Sprintf(`[Interface]
Address = %s
PrivateKey = %s
DNS = 8.8.8.8

[Peer]
Endpoint = %s
PublicKey = %s
AllowedIPs = 0.0.0.0/0
    `, client.AddressClient, client.PrivateClientKey, wg.Endpoint, wg.PublicKey)

	client.Config = clientConfig
	return client, clientID
}

// Отправка конфигурации через Telegram
func (wg *WireGuardConfig) SendConfigToUserTg(user_id int) {
	//создание бота
	Cl, _ := wg.Clients[user_id]

	chatID := telebot.ChatID(int64(Cl.TgId))
	bot, err := telebot.NewBot(telebot.Settings{
		Token: wg.BotToken,
	})

	//файл с конфигураций
	reader := strings.NewReader(Cl.Config)
	// Создаем документ для отправки, передавая reader как содержимое файла
	document := &telebot.Document{
		File:     telebot.FromReader(reader), // Используем io.Reader
		FileName: "wgconf.conf",              // Указываем имя файла
		Caption:  "WireGuard Configuration",  // Опциональная подпись к файлу
	}
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	//отправка файла
	if _, err := bot.Send(chatID, document); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
}

func (wg *WireGuardConfig) DropWireguard() {
	// очистка папки
	cmd := exec.Command("rm", "-rf", "/etc/wireguard/*")
	cmd.Run()
	err := cmd.Err
	if err != nil {
		log.Printf("failed to create keys : %v", err.Error())
	}
	//отключение wireguard
	cmd = exec.Command("systemctl", "disable", "wg-quick@wg0.service")
	cmd.Run()
	err = cmd.Err
	if err != nil {
		log.Printf("failed to create keys : %v", err.Error())
	}
	filePath := "/etc/sysctl.conf"

	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Читаем содержимое файла построчно
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Добавляем в список все строки, кроме той, которую нужно удалить
		if !strings.Contains(line, "net.ipv4.ip_forward=1") {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	// Перезаписываем файл без нужной строки
	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		log.Fatalf("failed to write file: %v", err)
	}

	fmt.Println("Line removed successfully")
	log.Printf("Папка конфиураций wireguuard очищена")
}

// // Сбор трафика
//
//	func (wg *WireGuardConfig) CollectTraffic() {
//		cmd := exec.Command("wg-json")
//		go cmd.Run() // Запускаем в горутине
//		log.Println("Сбор трафика начат. Для остановки используйте Ctrl+C.")
//	}
//
// Структура для хранения выходных данных команды wg-json
type PeerStats struct {
	TransferRx uint64
	TransferTx uint64
}

type PeerTraffic struct {
	Time    string `json:"time"`
	Traffic uint64 `json:"traffic"`
}

func (wg *WireGuardConfig) CollectTraffic() {
	previousPeerStats := make(map[string]PeerStats)
	for {
		// Run 'wg show all dump'
		cmd := exec.Command("wg", "show", "all", "dump")
		output, err := cmd.Output()

		if err != nil {
			log.Printf("Error running wg command: %v", err)
			continue
		}

		// Split output into lines
		lines := strings.Split(string(output), "\n")

		currentPeerStats := make(map[string]PeerStats)
		peerTraffic := make(map[string]PeerTraffic)
		currentTime := time.Now().Format(time.RFC3339)

		// Process each line
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Split(line, "\t")
			if len(fields) == 9 {
				// This is a peer line
				// Fields: interface, public-key, preshared-key, endpoint, allowed-ips, latest-handshake, transfer-rx, transfer-tx, persistent-keepalive
				peerAddress := fields[3] // Using 'endpoint' as the peer address
				transferRxStr := fields[6]
				transferTxStr := fields[7]

				transferRx, err := strconv.ParseUint(transferRxStr, 10, 64)
				if err != nil {
					log.Printf("Error parsing transferRx: %v", err)
					continue
				}
				transferTx, err := strconv.ParseUint(transferTxStr, 10, 64)
				if err != nil {
					log.Printf("Error parsing transferTx: %v", err)
					continue
				}

				currentPeerStats[peerAddress] = PeerStats{
					TransferRx: transferRx,
					TransferTx: transferTx,
				}

				// Get previous stats
				prevStats, exists := previousPeerStats[peerAddress]
				if exists {
					// Calculate traffic difference
					traffic := (transferRx - prevStats.TransferRx) + (transferTx - prevStats.TransferTx)
					peerTraffic[peerAddress] = PeerTraffic{
						Time:    currentTime,
						Traffic: traffic,
					}
				} else {
					// No previous stats, cannot calculate traffic
					peerTraffic[peerAddress] = PeerTraffic{
						Time:    currentTime,
						Traffic: 0,
					}
				}

			}
			// else if len(fields) == 5 {
			// Interface line, can skip
			// }
		}

		// Update previousPeerStats
		previousPeerStats = currentPeerStats

		// Create JSON data
		jsonData, err := json.MarshalIndent(peerTraffic, "", "  ")
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}

		// Save JSON to file
		err = ioutil.WriteFile("traffic.json", jsonData, 0644)
		if err != nil {
			log.Printf("Error writing JSON to file: %v", err)
			continue
		}
		// Sleep for 1 hour
		time.Sleep(3 * time.Minute)
	}
}
