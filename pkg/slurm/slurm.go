package slurm

type Client interface {
	GetJobs() (*JobsResponse, error)
	GetJob(id int) (*JobResponseProperties, error)
}
