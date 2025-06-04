package webserver

import (
	"encoding/json"
	"log"
	"os"
)

type Settings struct {
	UseHttps         bool
	UseHttpRedirect  bool
	Domain         string
	Bind string
	HttpPort         string
	HttpsPort        string
	Root             string
	FallbackRedirect string
	Logger           *log.Logger
	CertFile         string
	KeyFile          string
}

func NewSettings() *Settings {
	return &Settings{
		UseHttps:         false,
		UseHttpRedirect:  false,
		Domain:         "localhost",
Bind: "0,0,0,0",
		HttpPort:         "80",
		HttpsPort:        "443",
		Root:             "/",
		FallbackRedirect: "/404",
		Logger:           log.New(os.Stdout, "", log.LstdFlags),
		CertFile:         "",
		KeyFile:          "",
	}
}

func (s *Settings) SaveJson(fileName string) error {
	data, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (s *Settings) LoadJson(fileName string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return nil
}

func (s *Settings) Scheme() string {
	if s.UseHttps {
		return "https://"
	}
	return "http://"
}

func (s *Settings) Port() string {
	if s.UseHttps {
		return s.HttpsPort
	}
	return s.HttpPort
}

func (s *Settings) Addr() string {
	return s.Domain + ":" + s.Port()
}

func (s *Settings) BindAddr() string {
	return s.Bind + ":" + s.Port()
}

func (s *Settings) Url() string {
	return s.Scheme() + s.Addr()
}

func (s *Settings) UrlHttps() string {
	return "https://" + s.Domain + ":" + s.HttpsPort
}

func (s *Settings) UrlHttp() string {
	return "http://" + s.Domain + ":" + s.HttpPort
}
