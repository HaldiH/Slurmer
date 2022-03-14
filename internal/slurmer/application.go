package slurmer

type Application struct {
	Directory string
	Jobs      map[string]Job
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
