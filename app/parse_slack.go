package main

import (
	"encoding/json"
	"github.com/kudrykv/services-deploy-monitor/app/service"
	"io/ioutil"
)

func ParseSlack(ptf string) map[string]service.Slack {
	bts, err := ioutil.ReadFile(ptf)
	if err != nil {
		panic(err)
	}

	var jsonSlacks map[string]JsonSlack
	if json.Unmarshal(bts, &jsonSlacks); err != nil {
		panic(err)
	}

	slacks := map[string]service.Slack{}
	for key, jsonSlack := range jsonSlacks {
		slacks[key] = service.NewSlack(jsonSlack.Url)
	}

	return slacks
}
