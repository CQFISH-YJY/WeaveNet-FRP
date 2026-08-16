package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"weavenet/panel/config"
)

// mailItem 待发送邮件。
type mailItem struct {
	to      string
	subject string
	body    string
}

// mailQueue 内存邮件队列。
var mailQueue = make(chan mailItem, 1000)
var mailStarted = false

// StartMailWorker 启动邮件消费 worker。
func StartMailWorker() {
	if mailStarted {
		return
	}
	mailStarted = true
	go func() {
		for item := range mailQueue {
			for attempt := 0; attempt < 3; attempt++ {
				if err := sendMail(item); err == nil {
					break
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()
	log.Println("[mail] 邮件 worker 已启动")
}

// StopMailWorker 停止邮件 worker（关闭队列）。
func StopMailWorker() {
	if mailStarted {
		mailStarted = false
		close(mailQueue)
	}
}

// smtpConfigured 是否配置了 SMTP。
func smtpConfigured() bool {
	return config.C.SmtpHost != "" && config.C.SmtpFrom != ""
}

// sendMail 发送一封邮件（同步）。
func sendMail(item mailItem) error {
	if !smtpConfigured() {
		log.Printf("[邮件占位] to=%s subject=%s body=%s", item.to, item.subject, item.body)
		return nil
	}
	host := config.C.SmtpHost
	port := config.C.SmtpPort
	addr := fmt.Sprintf("%s:%d", host, port)
	from := config.C.SmtpFrom
	user := config.C.SmtpUser
	pass := config.C.SmtpPassword

	msg := "From: " + config.C.AppName + " <" + from + ">\r\n" +
		"To: " + item.to + "\r\n" +
		"Subject: " + item.subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		item.body

	var conn *smtp.Client
	var err error
	if config.C.SmtpUseSSL {
		tlsCfg := &tls.Config{ServerName: host, InsecureSkipVerify: false}
		tconn, terr := tls.Dial("tcp", addr, tlsCfg)
		if terr != nil {
			return terr
		}
		conn, err = smtp.NewClient(tconn, host)
	} else {
		conn, err = smtp.Dial(addr)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if config.C.SmtpUseTLS && !config.C.SmtpUseSSL {
		if err := conn.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}
	if user != "" {
		if err := conn.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
			return err
		}
	}
	if err := conn.Mail(from); err != nil {
		return err
	}
	if err := conn.Rcpt(item.to); err != nil {
		return err
	}
	w, err := conn.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

// EnqueueMail 入队邮件（队列满则丢弃）。
func EnqueueMail(to, subject, body string) {
	select {
	case mailQueue <- mailItem{to: to, subject: subject, body: body}:
	default:
		log.Printf("[mail] 邮件队列已满，丢弃: %s", subject)
	}
}

// SendVerificationCode 发送验证码邮件。
func SendVerificationCode(to, code, purpose string) {
	purposeText := map[string]string{
		"register":       "账号注册激活",
		"reset_password": "找回密码",
		"change_email":   "修改邮箱",
	}[purpose]
	if purposeText == "" {
		purposeText = purpose
	}
	subject := fmt.Sprintf("【%s】%s验证码", config.C.AppName, purposeText)
	body := fmt.Sprintf("您好，您正在%s。\n\n验证码：%s\n\n验证码 5 分钟内有效，请勿泄露给他人。\n\n如非本人操作，请忽略此邮件。", purposeText, code)
	EnqueueMail(to, subject, body)
}

// FormatSMTPConfig 将配置写入 system_configs（占位，供 admin 读取）。
func FormatSMTPConfig() map[string]any {
	return map[string]any{
		"smtp_host":     config.C.SmtpHost,
		"smtp_port":     config.C.SmtpPort,
		"smtp_user":     config.C.SmtpUser,
		"smtp_from":     config.C.SmtpFrom,
		"smtp_use_ssl":  config.C.SmtpUseSSL,
		"smtp_use_tls":  config.C.SmtpUseTLS,
		"smtp_password": strings.Repeat("*", len(config.C.SmtpPassword)),
	}
}
