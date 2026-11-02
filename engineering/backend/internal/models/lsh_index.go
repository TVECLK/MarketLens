package models

// LSH dedup index

type LshIndex struct {
	BucketKey string   `json:"bucket_key" gorm:"primaryKey;size:64;not null"`
	BandNo    int      `json:"band_no"    gorm:"not null"`
	JobPostID uint     `json:"job_post_id" gorm:"primaryKey;not null"`
	JobPost   *JobPost `json:"-"          gorm:"foreignKey:JobPostID;constraint:OnDelete:CASCADE;"`
}

func (LshIndex) TableName() string { return "lsh_index" }
