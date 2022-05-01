package persistent

import (
	"encoding/json"

	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dbJobs struct {
	db *gorm.DB
}

func NewPersistentJobs(db *gorm.DB) (containers.JobsContainer, error) {
	if err := db.AutoMigrate(&model.Job{}); err != nil {
		return nil, err
	}
	return &dbJobs{db: db}, nil
}

func (jobsContainer *dbJobs) GetAllJobs() ([]*model.Job, error) {
	var jobs []*model.Job
	if res := jobsContainer.db.Preload("SlurmJob").Find(&jobs); res.Error != nil {
		return nil, res.Error
	}
	return jobs, nil
}

func (jobsContainer *dbJobs) GetJob(jobId uuid.UUID) (*model.Job, error) {
	var job *model.Job
	if res := jobsContainer.db.Preload("SlurmJob").First(job, jobId); res.Error != nil {
		return nil, res.Error
	}
	return job, nil
}

func (jobsContainer *dbJobs) DeleteJob(jobId uuid.UUID) error {
	return jobsContainer.db.
		Select("SlurmJob").
		Delete(&model.Job{}, jobId).
		Error
}

func (jobsContainer *dbJobs) UpdateJob(job *model.Job) error {
	return jobsContainer.db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Save(job).
		Error
}

func (jobsContainer *dbJobs) GetAllAppJobs(appId uuid.UUID) ([]*model.Job, error) {
	var jobs []*model.Job
	if res := jobsContainer.db.Preload("SlurmJob").Where("app_id = ?", appId).Find(&jobs); res.Error != nil {
		return nil, res.Error
	}
	return jobs, nil
}

func (jobsContainer *dbJobs) GetAppJob(appId uuid.UUID, jobId uuid.UUID) (*model.Job, error) {
	var job model.Job
	if res := jobsContainer.db.Preload("SlurmJob").Where("app_id = ? AND id = ?", appId, jobId).First(&job); res.Error != nil {
		return nil, res.Error
	}
	return &job, nil
}

func (jobsContainer *dbJobs) AddAppJob(appId uuid.UUID, job *model.Job) error {
	job.AppId = appId
	return jobsContainer.db.Create(job).Error
}

func (jobsContainer *dbJobs) DeleteAppJob(appId uuid.UUID, jobId uuid.UUID) error {
	return jobsContainer.db.
		Select(clause.Associations).
		Where("app_id = ?", appId).
		Delete(&model.Job{}, jobId).
		Error
}

func (jobsContainer *dbJobs) UpdateAppJob(appId uuid.UUID, job *model.Job) error {
	return jobsContainer.db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Where("app_id = ?", appId).
		Save(job).
		Error
}

func (c *dbJobs) MarshalJSON() ([]byte, error) {
	jobs, err := c.GetAllJobs()
	if err != nil {
		return nil, err
	}
	return json.Marshal(jobs)
}
