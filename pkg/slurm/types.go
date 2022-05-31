package slurm

type MetaProperties struct {
	Plugin struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"plugin"`
	Slurm struct {
		Version struct {
			Major int `json:"major"`
			Micro int `json:"micro"`
			Minor int `json:"minor"`
		} `json:"version"`
		Release string `json:"release"`
	} `json:"Slurm"`
}

type Error struct {
	Error *string `json:"error"`
	Errno *int    `json:"errno"`
}

type JobResponseProperties struct {
	Account          string   `json:"account"`
	JobId            int      `json:"job_id" gorm:"primaryKey"`
	JobState         JobState `json:"job_state"`
	StateDescription string   `json:"state_description"`
	StateReason      string   `json:"state_reason"`
	Name             string   `json:"name"`
	UserName         string   `json:"user_name"`
	UserId           int      `json:"user_id"`
}

type JobsResponse struct {
	Errors []Error                 `json:"errors"`
	Jobs   []JobResponseProperties `json:"jobs"`
	Meta   MetaProperties          `json:"meta"`
}

type BatchProperties struct {
	Account     string   `json:"account" validate:"excludesrune=\n"`
	Chdir       string   `json:"chdir" validate:"excludesrune=\n"`
	Comment     string   `json:"comment" validate:"excludesrune=\n"`
	CpusPerTask uint     `json:"cpus_per_task"`
	JobName     string   `json:"job_name" validate:"excludesrune=\n"`
	Command     string   `json:"command" validate:"required"`
	Args        []string `json:"args"`
}

type SBatchOptions struct {
	Array   []int  `json:"array"`
	Account string `json:"account"`
	Begin   string `json:"begin"`
	Wait    bool   `json:"wait"`
	Chdir   string `json:"chdir"`
	Uid     string `json:"uid"`
	Gid     string `json:"gid"`
}

type JobState string

const (
	BOOT_FAIL     JobState = "BOOT_FAIL"
	CANCELLED     JobState = "CANCELLED"
	COMPLETED     JobState = "COMPLETED"
	CONFIGURING   JobState = "CONFIGURING"
	COMPLETING    JobState = "COMPLETING"
	DEADLINE      JobState = "DEADLINE"
	FAILED        JobState = "FAILED"
	NODE_FAIL     JobState = "NODE_FAIL"
	OUT_OF_MEMORY JobState = "OUT_OF_MEMORY"
	PENDING       JobState = "PENDING"
	PREEMPTED     JobState = "PREEMPTED"
	RUNNING       JobState = "RUNNING"
	RESV_DEL_HOLD JobState = "RESV_DEL_HOLD"
	REQUEUE_FED   JobState = "REQUEUE_FED"
	REQUEUE_HOLD  JobState = "REQUEUE_HOLD"
	REQUEUED      JobState = "REQUEUED"
	RESIZING      JobState = "RESIZING"
	REVOKED       JobState = "REVOKED"
	SIGNALING     JobState = "SIGNALING"
	SPECIAL_EXIT  JobState = "SPECIAL_EXIT"
	STAGE_OUT     JobState = "STAGE_OUT"
	STOPPED       JobState = "STOPPED"
	SUSPENDED     JobState = "SUSPENDED"
	TIMEOUT       JobState = "TIMEOUT"
)

func (s JobState) IsStopped() bool {
	switch s {
	case COMPLETED,
		CANCELLED,
		BOOT_FAIL,
		DEADLINE,
		FAILED,
		NODE_FAIL,
		PREEMPTED,
		TIMEOUT:
		return true
	default:
		return false
	}
}
