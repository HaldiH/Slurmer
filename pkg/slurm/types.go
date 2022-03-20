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
	Account          *string `json:"account"`
	JobId            *int    `json:"job_id"`
	JobState         *string `json:"job_state"`
	StateDescription *string `json:"state_description"`
	StateReason      *string `json:"state_reason"`
	Name             *string `json:"name"`
}

type JobsResponse struct {
	Errors []Error                 `json:"errors"`
	Jobs   []JobResponseProperties `json:"jobs"`
	Meta   MetaProperties          `json:"meta"`
}
