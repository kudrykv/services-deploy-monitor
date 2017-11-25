package config

type Config struct {
	Server   Server
	Github   Github
	CircleCi CircleCi
	Monitor  Monitor
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

type Monitor struct {
	PollTimeIntervalS int `env:"POLL_TIME_INTERVAL_SECONDS" envDefault:"10"`
	// PollForBuildsTimes defines how many times search for for build in the Ci response.
	PollForBuildsTimes      int `env:"POLL_FOR_BUILDS_TIMES" envDefault:"3"`
	PollForGreenBuildsTimes int `env:"POLL_FOR_GREEN_BUILDS_TIMES" envDefault:"20"`
}
