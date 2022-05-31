package slurm

type Client interface {
	GetJobs(ids ...int) ([]JobResponseProperties, error)
	GetJob(id int) (*JobResponseProperties, error)
	SubmitJob(user string, o *SBatchOptions, script string, cwd string) (jobID int, err error)
	CancelJob(user string, ids ...int) error
}
