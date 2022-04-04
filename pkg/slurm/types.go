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
	Account          string `json:"account"`
	JobId            int    `json:"job_id" gorm:"primaryKey"`
	JobState         string `json:"job_state"`
	StateDescription string `json:"state_description"`
	StateReason      string `json:"state_reason"`
	Name             string `json:"name"`
}

type JobsResponse struct {
	Errors []Error                 `json:"errors"`
	Jobs   []JobResponseProperties `json:"jobs"`
	Meta   MetaProperties          `json:"meta"`
}

type BatchProperties struct {
	Account     string   `json:"account"`
	Chdir       string   `json:"chdir"`
	Comment     string   `json:"comment"`
	CpusPerTask uint     `json:"cpus_per_task"`
	JobName     string   `json:"job_name"`
	Command     string   `json:"command"`
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
