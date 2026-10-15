package models

// API request payloads

type BatchSavePayload struct {
	NewJobs    []JobPost  `json:"new_jobs"    binding:"required"`
	LshIndexes []LshIndex `json:"lsh_indexes" binding:"required"`
}

type BatchUpdatePayload struct {
	Duplicates []JobMetaData `json:"duplicates" binding:"required"`
}

type ReconciliationPayload struct {
	CrawlerRunID uint `json:"crawler_run_id" binding:"required"`
}
