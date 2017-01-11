package config

import (
	"html/template"
	"os"

	"github.com/Sky-And-Hammer/render"
	//	Golang Configuration tool that support YAML, JSON, Shell Environment
	"github.com/jinzhu/configor"
	//	a HTML sanitizer implemented in Go. It is fast and highly configurable.
	"github.com/microcosm-cc/bluemonday"
)

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Site     string
}

var Config = struct {
	Port uint `defualt:"7000" env:"PORT"`
	DB   struct {
		Name     string `default:"qor_example"`
		Adapter  string `default:"mysql"`
		User     string
		Password string
	}
	STMP SMTPConfig
}{}

var (
	Root = os.Getenv("GOPATH") + "/src/github.com/"
	View *render.Render
)

func init() {
	if err := configor.Load(&Config, "config/database.yml", "config/smtp.yml"); err != nil {
		panic(err)
	}

	View = render.New()
	htmlSanitizer := bluemonday.UGCPolicy()
	View.RegisterFuncMap("raw", func(str string) template.HTML {
		return template.HTML(htmlSanitizer.Sanitize(str))
	})
}

func (s SMTPConfig) HostWithPort() string {
	return s.Host + ":" + s.Port
}
