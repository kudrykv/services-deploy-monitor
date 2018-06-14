package main

type jsonConfig struct {
	Cvs JsonCvs `json:"cvs"`
}

type JsonCvs struct {
	Branches map[string]JsonCvsItem `json:"branches"`
}

type JsonCvsItem struct {
	Github   map[string]JsonSystems            `json:"github"`
	CircleCi map[string]map[string]JsonSystems `json:"circle_ci"`
}

type JsonSystems struct {
	Slack   string `json:"slack"`
	Message string `json:"message"`
}
