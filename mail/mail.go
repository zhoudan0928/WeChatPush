package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	from       mail.Address
	to         mail.Address
	smtpServer string
	smtpPort   string
	username   string
	password   string
)

func init() {
	// 尝试在不同位置加载 .env 文件
	envPaths := []string{".env", "../.env", "../../.env", "/app/.env"}
	envLoaded := false

	for _, path := range envPaths {
		absPath, _ := filepath.Abs(path)
		if err := godotenv.Load(absPath); err == nil {
			log.Printf("成功加载 .env 文件: %s", absPath)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("警告: 无法加载 .env 文件，将使用环境变量")
	}

	from = mail.Address{
		Name:    getEnv("FROM_NAME", "发件人"),
		Address: getEnv("FROM_ADDRESS", ""),
	}
	to = mail.Address{
		Name:    getEnv("TO_NAME", "收件人"),
		Address: getEnv("TO_ADDRESS", ""),
	}
	smtpServer = getEnv("SMTP_SERVER", "")
	smtpPort = getEnv("SMTP_PORT", "587")
	username = getEnv("USERNAME", "")
	password = getEnv("PASSWORD", "")

	if from.Address == "" || to.Address == "" || smtpServer == "" || username == "" || password == "" {
		log.Println("警告: 一些必要的环境变量未设置")
		log.Printf("FROM_ADDRESS: %s", from.Address)
		log.Printf("TO_ADDRESS: %s", to.Address)
		log.Printf("SMTP_SERVER: %s", smtpServer)
		log.Printf("USERNAME: %s", username)
		log.Printf("PASSWORD: %s", password != "")
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func SendEmail(name string, content string) error {
	auth := smtp.PlainAuth("", username, password, smtpServer)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}

	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("无法连接到SMTP服务器: %v", err)
	}
	defer conn.Close()

	smtpClient, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return fmt.Errorf("无法创建SMTP客户端: %v", err)
	}
	defer smtpClient.Quit()

	if err = smtpClient.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	if err = smtpClient.Mail(from.Address); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	if err = smtpClient.Rcpt(to.Address); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("创建邮件数据写入器失败: %v", err)
	}
	defer w.Close()

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from.String(), to.String(), name, content)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	log.Println("邮件发送成功")
	return nil
}
