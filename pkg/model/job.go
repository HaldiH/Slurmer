package model

import (
	"encoding/json"
	"errors"

	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"github.com/google/uuid"
)

type Job struct {
	Name      string                       `json:"name"`
	Status    JobStatus                    `json:"status"`
	UserName  string                       `json:"user_name"`
	Id        uuid.UUID                    `json:"id" gorm:"primaryKey"`
	SlurmId   int                          `json:"slurm_id"`
	SlurmJob  *slurm.JobResponseProperties `json:"slurm_job" gorm:"foreignKey:SlurmId"`
	Directory string                       `json:"-"`
	AppId     uuid.UUID                    `json:"-"`
}

type JobPatchRequest struct {
	Action *JobAction `json:"action"`
}

type JobAction string

const (
	JobPrune JobAction = "prune"
	JobStart JobAction = "start"
	JobStop  JobAction = "stop"
)

var ErrUnknownJobStatus error = errors.New("unknown job status")

type JobStatus uint

const (
	JobStopped JobStatus = iota
	JobStarted
)

func (s JobStatus) String() string {
	switch s {
	case JobStopped:
		return "stopped"
	case JobStarted:
		return "started"
	}
	return "unknown"
}

func (s JobStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *JobStatus) UnmarshalJSON(data []byte) error {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	switch status {
	case "stopped":
		*s = JobStopped
	case "started":
		*s = JobStarted
	default:
		return ErrUnknownJobStatus
	}
	return nil
}
