package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"os"
	"os/exec"
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

// Конфигурационная клиента
type Client struct {
	Id               int
	Satus            bool
	AddressClient    string
	Pubkey_path      string
	Privkey_path     string
	PrivateClientKey string
	PublickClientKey string
	Peer             PeerConfig
	Peer_str         string
	Config           string
	Tg_id            int
}

// Управление сервером wiredurd
type WireGuardConfig struct {
	PrivateKey string
	PublicKey  string
	Endpoint   string
	ListenPort string
	InterName  string
	BotToken   string

	Clients map[int]*Client
}

// ------------------------ методы для клиентов ------------------------
// Остановка клиента
func (clients *WireGuardConfig) StopClient(id int) {
	client := clients.Clients[id]
	filePath := "/etc/wireguard/wg0.conf"
	// Чтение содержимого файла
	content, _ := os.ReadFile(filePath)
	// Преобразование содержимого в строку
	fileContent := string(content)
	// Поиск и удаление блока Peer
	updatedContent := strings.Replace(fileContent, client.Peer_str, "", 1)
	// Перезапись файла
	os.WriteFile(filePath, []byte(updatedContent), 0644)
}

// активация клиента
func (clients *WireGuardConfig) ActClient(id int) {
	client := clients.Clients[id]
	filePath := "/etc/wireguard/wg0.conf"
	os.WriteFile(filePath, []byte(client.Peer_str), 0644)
}

// Удаление клиента
func (clients *WireGuardConfig) DeleteClient(id int) {
	clients.StopClient(id)
	client := clients.Clients[id]
	err := os.Remove(client.Privkey_path)
	if err != nil {
		log.Printf("не удалось удалить файл: %v", err)
	}
	err = os.Remove(client.Pubkey_path)
	if err != nil {
		log.Printf("не удалось удалить файл: %v", err)
	}
}

// вывод всех клиентов
func (clients *WireGuardConfig) AllClients() string {
	text := ""
	for _, client := range clients.Clients {
		var stat string
		if client.Satus {
			stat = "Остановлен"
		} else {
			stat = "Активен"
		}
		text += fmt.Sprintf("Клиент %d статус %s адресс %s \n", client.Id, stat, client.AddressClient)
	}
	return text

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

// Генерация конфигурации клиента WireGuard
func (wg *WireGuardConfig) AddWireguardClient(client_id int) (string, *Client) {
	i, ok := wg.Clients[client_id]
	Config := &Client{}
	if ok {
		Config = i
	} else {
		Config = &Client{}
		wg.Clients[client_id] = Config
	}
	tmpl := `[Interface]
Address = {{.AddressClient}}
PrivateKey = {{.PrivateClientKey}}
DNS = 8.8.8.8

[Peer]
Endpoint = {{.Endpoint}}
PublicKey = {{.PublicKey}}
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 10`
	// имена ключей
	clientPubname := fmt.Sprintf("/etc/wireguard/wg_client_%d_punlick", client_id)
	clientPrivname := fmt.Sprintf("/etc/wireguard/wg_client_%d_private", client_id)
	var clienPrivateKey, clienPublickKey string
	////
	var privateKey bytes.Buffer
	cmd := exec.Command("wg", "genkey")
	cmd.Stdout = &privateKey
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}
	// Сохраняем приватный ключ в переменную
	clienPrivateKey = strings.ReplaceAll(privateKey.String(), "\n", "")
	// Используем приватный ключ для генерации публичного ключа
	var publicKey bytes.Buffer
	cmd = exec.Command("wg", "pubkey")
	cmd.Stdin = &privateKey
	cmd.Stdout = &publicKey

	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to generate public key: %v", err)
	}
	clienPublickKey = strings.ReplaceAll(publicKey.String(), "\n", "")
	//запись
	os.WriteFile(clientPrivname, []byte(clienPrivateKey), 0600)
	os.WriteFile(clientPubname, []byte(clienPublickKey), 0600)
	//
	log.Printf("Ключи для клента созданы и записаны в файлы")
	fmt.Println("приватный ключ : ", clienPrivateKey)
	fmt.Println("публичный ключ : ", clienPublickKey)

	//доностройка конфига клиента
	//config.ServerPubKey = publickkey
	Config.PrivateClientKey = clienPrivateKey
	Config.PublickClientKey = clienPublickKey
	Config.AddressClient = fmt.Sprintf("10.0.0.%d/24", client_id)
	// файл конфигурации
	file, err := os.OpenFile("/etc/wireguard/wg0.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
	}
	defer file.Close()
	// добавление пира
	peer := `

[Peer]
PublicKey = {{.PublickClientKey}}
AllowedIPs = {{.AddressClient}}
`
	t_p := template.Must(template.New("wgClientConfig").Parse(peer))
	if err := t_p.Execute(file, Config); err != nil {
	}
	// создание клиентского конфига
	t := template.Must(template.New("wgClientConfig").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, Config); err != nil {
	}
	//перезапуск wireguard для сощ=здания пиров
	cmd = exec.Command("sudo", "systemctl", "restart", "wg-quick@wg0")
	cmd.Run()
	err = cmd.Err
	if err != nil {
		log.Printf("failed to create keys : %v", err.Error())
	}
	log.Printf("wireguard перезапущен")
	Config.Config = buf.String()
	return buf.String(), Config
}

// Отправка конфигурации через Telegram
func (wg *WireGuardConfig) SendConfigToUserTg(user_id int) {
	//создание бота
	i, ok := wg.Clients[user_id]
	Cl := &Client{}
	if ok {
		Cl = i
	} else {
		Cl = &Client{}
		wg.Clients[user_id] = Cl
	}
	chatID := telebot.ChatID(int64(Cl.Tg_id))
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

// Сбор трафика
func (wg *WireGuardConfig) CollectTraffic() {
	cmd := exec.Command("tcpdump", "-i", wg.InterName, "-w", "traffic.pcap", "-T", "json", ">", "traffic.json")
	go cmd.Run() // Запускаем в горутине
	log.Println("Сбор трафика начат. Для остановки используйте Ctrl+C.")
}
