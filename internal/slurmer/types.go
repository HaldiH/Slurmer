package slurmer

type BatchProperties struct {
	Account     string `json:"account"`
	Chdir       string `json:"chdir"`
	Comment     string `json:"comment"`
	CpusPerTask uint   `json:"cpus_per_task"`
	JobName     string `json:"job_name"`
}
