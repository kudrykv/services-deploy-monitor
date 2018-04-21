package notifier

type Event struct {
	Action    string
	SubAction string

	Repo Repo
	Pr   *Pr
}

type Repo struct {
	Org    string
	Name   string
	Branch string
	Tag    string
	Sha    string
}

type Pr struct {
	Base   string
	Title  string
	Number int
}
