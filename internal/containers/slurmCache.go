package containers

import "github.com/ShinoYasx/Slurmer/pkg/slurm"

type SlurmCache interface {
	SetSlurmJob(slurmJob *slurm.JobResponseProperties) error
	GetSlurmJob(slurmJobId int) (*slurm.JobResponseProperties, error)
	DeleteSlurmJob(id int) error
}
