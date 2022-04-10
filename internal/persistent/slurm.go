package persistent

import (
	"github.com/ShinoYasx/Slurmer/internal/containers"
	"github.com/ShinoYasx/Slurmer/pkg/slurm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type slurmCache struct {
	db *gorm.DB
}

func NewSlurmCache(db *gorm.DB) (containers.SlurmCache, error) {
	if err := db.AutoMigrate(&slurm.JobResponseProperties{}); err != nil {
		return nil, err
	}
	if res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&slurm.JobResponseProperties{}); res.Error != nil {
		return nil, res.Error
	}
	return &slurmCache{db: db}, nil
}

func (c *slurmCache) SetSlurmJob(slurmJob *slurm.JobResponseProperties) error {
	return c.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(slurmJob).Error
}

func (c *slurmCache) GetSlurmJob(slurmJobId int) (*slurm.JobResponseProperties, error) {
	var job slurm.JobResponseProperties
	if res := c.db.Where("job_id = ?", slurmJobId).First(&job); res.Error != nil {
		return nil, res.Error
	}
	return &job, nil
}

func (c *slurmCache) DeleteSlurmJob(id int) error {
	return c.db.Delete(&slurm.JobResponseProperties{}, id).Error
}
