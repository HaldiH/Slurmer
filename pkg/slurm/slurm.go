package slurm

type Client interface {
	GetJobs(ids ...int) ([]JobResponseProperties, error)
	GetJob(id int) (*JobResponseProperties, error)
	SubmitJob(o *SBatchOptions, script string, cwd string) (jobID int, err error)
	CancelJob(id int) error
}
