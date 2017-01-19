package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Server Server    `yaml:"server"`
	From   string    `yaml:"from"`
	Auth   PlainAuth `yaml:"auth"`
}

type Server struct {
	Host      string `yaml:"host"`
	TLS       bool   `yaml:"tls"`
	VerifyTLS bool   `yaml:"verify-tls"`
}

type PlainAuth struct {
	Username string
	Password string
}

const DefaultConfigurationFile = "/etc/go-smtp-send.yaml"

func main() {
	var config Configuration
	log.SetFlags(0)

	help := flag.Bool("h", false, "show this help")
	configFile := flag.String("config", DefaultConfigurationFile, "config file to use")
	to := flag.String("to", "", "address to send mail to")
	subj := flag.String("subject", "", "subject of the email")

	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *to == "" {
		log.Fatalln("error: must specify to (-to)")
	}

	configBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalln("error: could not read config:", err)
	}

	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		log.Fatalln("error: could not read config:", err)
	}

	if err := CheckConfig(config); err != nil {
		log.Fatalln("error: could not read config:", err)
	}

	cl, err := Connect(config)
	if err != nil {
		log.Fatalln("error: could not connect:", err)
	}

	if err := cl.Mail(config.From); err != nil {
		log.Fatalln("error: could not MAIL:", err)
	}
	if err := cl.Rcpt(*to); err != nil {
		log.Fatalln("error: could not RCPT:", err)
	}

	w, err := cl.Data()
	if err != nil {
		log.Fatalln("error: could not DATA:", err)
	}

	headers := []string{
		fmt.Sprintf("From: %s\r\n", config.From),
		fmt.Sprintf("To: %s\r\n", to),
		fmt.Sprintf("Subject: %s\r\n", *subj),
	}
	for _, h := range headers {
		if _, err := w.Write([]byte(h)); err != nil {
			log.Fatalln("error: could not send body:", err)
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		log.Fatalln("error: could not send body:", err)
	}

	if _, err := io.Copy(w, os.Stdin); err != nil {
		log.Fatalln("error: could not send body:", err)
	}

	if err := w.Close(); err != nil {
		log.Fatalln("error: could not send body:", err)
	}

	cl.Quit()
}

func CheckConfig(config Configuration) error {
	if config.From == "" {
		return fmt.Errorf("from field empty")
	}
	if config.Server.Host == "" {
		return fmt.Errorf("no host specified")
	}
	return nil
}

func Connect(config Configuration) (*smtp.Client, error) {
	host, _, err := net.SplitHostPort(config.Server.Host)
	if err != nil {
		return nil, fmt.Errorf("could not read host: %s", err)
	}

	var con net.Conn
	if config.Server.TLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: !config.Server.VerifyTLS,
			ServerName:         host,
		}

		con, err = tls.Dial("tcp", config.Server.Host, tlsconfig)
	} else {
		con, err = net.Dial("tcp", config.Server.Host)
	}
	if err != nil {
		return nil, err
	}

	cl, err := smtp.NewClient(con, host)
	if err != nil {
		return nil, err
	}

	if config.Auth.Username != "" {
		auth := smtp.PlainAuth(
			"",
			config.Auth.Username,
			config.Auth.Password,
			host,
		)
		if err := cl.Auth(auth); err != nil {
			return nil, err
		}
	}

	return cl, nil
}
