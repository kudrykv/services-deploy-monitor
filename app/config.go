package main

type Config struct {
	Server   Server
	Github   Github
	CircleCi CircleCi
}

type Server struct {
	Port string `env:"PORT" envDefault:"8080"`
}

type Github struct {
	Key string `env:"GITHUB_KEY"`
	Org string `env:"GITHUB_ORG"`
}

type CircleCi struct {
	Key string `env:"CIRCLE_CI_KEY"`
	Org string `env:"CIRCLE_CI_ORG"`
}
