package notifier

import (
	"regexp"
	"text/template"
)

type Config struct {
	Cvs Cvs
}

type Cvs struct {
	Branches map[*regexp.Regexp]Systems
	Tags     map[*regexp.Regexp]Systems
}

type Systems struct {
	Github map[string]SendPack
}

type SendPack struct {
	Message *template.Template
}
