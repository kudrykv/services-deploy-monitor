package main

import (
	"encoding/json"
	"errors"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"io/ioutil"
	"regexp"
	"text/template"
)

func ParseConfig(ptf string, slacks map[string]service.Slack) service.Config {
	bts, err := ioutil.ReadFile(ptf)
	if err != nil {
		panic(err)
	}

	var jc jsonConfig
	if err := json.Unmarshal(bts, &jc); err != nil {
		panic(err)
	}

	config := service.Config{
		Cvs: service.Cvs{
			Branches: map[*regexp.Regexp]service.Systems{},
		},
	}

	for ptn, rest := range jc.Cvs.Branches {
		r := regexp.MustCompile(ptn)

		ss := service.Systems{
			Github:   map[string]service.SendPack{},
			CircleCi: map[string]map[string]service.SendPack{},
		}

		for event, smth := range rest.Github {
			tpl := template.New(ptn + event)
			parsed, err := tpl.Parse(smth.Message)
			if err != nil {
				panic(err)
			}

			slack, ok := slacks[smth.Slack]
			if !ok {
				panic(errors.New("slack " + smth.Slack + " has not been found"))
			}

			ss.Github[event] = service.SendPack{
				Message: parsed,
				Room:    smth.Room,
				Slack:   slack,
			}
		}

		for event, mapOfSystems := range rest.CircleCi {
			neededMap := map[string]service.SendPack{}

			for status, smth := range mapOfSystems {
				tpl := template.New(ptn + event + status)
				parsed, err := tpl.Parse(smth.Message)
				if err != nil {
					panic(err)
				}

				slack, ok := slacks[smth.Slack]
				if !ok {
					panic(errors.New("slack " + smth.Slack + " has not been found"))
				}

				neededMap[status] = service.SendPack{
					Message: parsed,
					Room:    smth.Room,
					Slack:   slack,
				}
			}

			ss.CircleCi[event] = neededMap
		}

		config.Cvs.Branches[r] = ss
	}

	return config
}
