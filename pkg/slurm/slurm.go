package slurm

type Client interface {
	GetJobs(ids ...int) (*JobsResponse, error)
	GetJob(id int) (*JobResponseProperties, error)
	SubmitBatch(o SBatchOptions) (jobID int, err error)
	CancelJob(id int) error
}
