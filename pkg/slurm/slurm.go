package slurm

type Client interface {
	GetJobs(ids ...int) (*JobsResponse, error)
	GetJob(id int) (*JobResponseProperties, error)
}
