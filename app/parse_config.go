package main

import (
	"encoding/json"
	"fmt"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"io/ioutil"
	"regexp"
	"text/template"
)

func ParseConfig(ptf string) service.Config {
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

			ss.Github[event] = service.SendPack{
				Message: parsed,
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

				neededMap[status] = service.SendPack{
					Message: parsed,
				}
			}

			ss.CircleCi[event] = neededMap
		}

		config.Cvs.Branches[r] = ss
	}

	fmt.Println(config.Cvs.Branches)

	return config
}
