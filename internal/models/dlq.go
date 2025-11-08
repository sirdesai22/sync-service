package models

import "time"

type DLQ struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	OutboxID   int64     `gorm:"index"`
	EntityType string
	EntityID   string
	Op         string
	ErrorMsg   string
	Payload    []byte    `gorm:"type:bytea"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	RetriedAt  *time.Time
	Resolved   bool `gorm:"default:false"`
}
