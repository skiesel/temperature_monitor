package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
)

type SMTP_Config struct {
	Server         string `json:"smtp_server"`
	Port           string `json:"smtp_port"`
	User           string `json:"smtp_user"`
	Password       string `json:"smtp_password"`
	ContactAddress string `json:"contact_address"`
}

var (
	settings = SMTP_Config{}
)

func init() {
	fileReader, err := os.Open("settings.config")

	decoder := json.NewDecoder(fileReader)
	err = decoder.Decode(&settings)
	if err != nil {
		panic(err)
	}
}

func SendMail(subject, body string) {
	from := mail.Address{Name: "",
		Address: settings.User,
	}
	to := mail.Address{Name: "",
		Address: settings.ContactAddress,
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	host := settings.Server
	port := settings.Port
	fullHost := host + ":" + port

	auth := smtp.PlainAuth("", settings.User, settings.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", fullHost, tlsconfig)
	if err != nil {
		panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		panic(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		panic(err)
	}

	// Data
	wr, err := c.Data()
	if err != nil {
		panic(err)
	}

	_, err = wr.Write([]byte(message))
	if err != nil {
		panic(err)
	}

	err = wr.Close()
	if err != nil {
		panic(err)
	}

	c.Quit()
}
