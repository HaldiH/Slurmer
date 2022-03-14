package slurmer

type JobsContainer map[string]*Job

type Application struct {
	Directory string
	Jobs      JobsContainer
}

const (
	Stopped int = iota
	Started
)

type Job struct {
	Name   string
	Status int
	PID    string
}
