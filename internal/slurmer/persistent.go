package slurmer

import (
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PersistentJobs struct {
	db *gorm.DB
}

func NewPersistentJobs(db *gorm.DB) (*PersistentJobs, error) {
	if err := db.AutoMigrate(&Job{}); err != nil {
		return nil, err
	}
	return &PersistentJobs{db: db}, nil
}

func (jobsContainer *PersistentJobs) GetAllJobs() ([]*Job, error) {
	var jobs []*Job
	if res := jobsContainer.db.Preload("SlurmJob").Find(&jobs); res.Error != nil {
		return nil, res.Error
	}
	return jobs, nil
}

func (jobsContainer *PersistentJobs) GetJob(jobID string) (*Job, error) {
	var job *Job
	if res := jobsContainer.db.Preload("SlurmJob").First(job, jobID); res.Error != nil {
		return nil, res.Error
	}
	return job, nil
}

func (jobsContainer *PersistentJobs) DeleteJob(jobId string) error {
	return jobsContainer.db.
		Select(clause.Associations).
		Where("job_id = ?", jobId).
		Delete(&Job{}).
		Error
}

func (jobsContainer *PersistentJobs) UpdateJob(job *Job) error {
	return jobsContainer.db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Save(job).
		Error
}

func (jobsContainer *PersistentJobs) GetAllAppJobs(appID string) ([]*Job, error) {
	var jobs []*Job
	if res := jobsContainer.db.Preload("SlurmJob").Where("app_id = ?", appID).Find(&jobs); res.Error != nil {
		return nil, res.Error
	}
	return jobs, nil
}

func (jobsContainer *PersistentJobs) GetAppJob(appID string, jobID string) (*Job, error) {
	var job Job
	if res := jobsContainer.db.Preload("SlurmJob").Where("app_id = ? AND id = ?", appID, jobID).First(&job); res.Error != nil {
		return nil, res.Error
	}
	return &job, nil
}

func (jobsContainer *PersistentJobs) AddAppJob(appId string, job *Job) error {
	job.AppId = appId
	return jobsContainer.db.Create(job).Error
}

func (jobsContainer *PersistentJobs) DeleteAppJob(appId string, jobId string) error {
	return jobsContainer.db.
		Select(clause.Associations).
		Where("app_id = ? AND id = ?", appId, jobId).
		Delete(&Job{}).
		Error
}

func (jobsContainer *PersistentJobs) UpdateAppJob(appID string, job *Job) error {
	return jobsContainer.db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Where("app_id = ?", appID).
		Save(job).
		Error
}

type SlurmCache struct {
	db *gorm.DB
}

func NewSlurmCache(db *gorm.DB) (*SlurmCache, error) {
	if err := db.AutoMigrate(&slurm.JobResponseProperties{}); err != nil {
		return nil, err
	}
	if res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&slurm.JobResponseProperties{}); res.Error != nil {
		return nil, res.Error
	}
	return &SlurmCache{db: db}, nil
}

func (c *SlurmCache) SetSlurmJob(slurmJob *slurm.JobResponseProperties) error {
	return c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(slurmJob).Error
}

func (c *SlurmCache) GetSlurmJob(slurmJobId int) (*slurm.JobResponseProperties, error) {
	var job slurm.JobResponseProperties
	if res := c.db.Where("job_id = ?", slurmJobId).First(&job); res.Error != nil {
		return nil, res.Error
	}
	return &job, nil
}

func (c *SlurmCache) DeleteSlurmJob(id int) error {
	return c.db.Delete(&slurm.JobResponseProperties{}, id).Error
}