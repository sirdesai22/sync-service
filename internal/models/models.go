package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ---------------- USERS ----------------
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username  string         `gorm:"uniqueIndex;not null"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Skills    datatypes.JSON // store []string as JSON
	College   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Projects  []Project `gorm:"foreignKey:OwnerID"`
}

// ---------------- HACKATHONS ----------------
type Hackathon struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"`
	Location  string
	StartAt   time.Time
	EndAt     time.Time
	Tracks    datatypes.JSON // e.g. ["AI","Web3"]
	CreatedAt time.Time
	UpdatedAt time.Time
	Projects  []Project `gorm:"foreignKey:HackathonID"`
}

// ---------------- PROJECTS ----------------
type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"index;not null"`
	Description string
	HackathonID uuid.UUID
	OwnerID     uuid.UUID
	TeamMembers datatypes.JSON // store []uuid
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ---------------- OUTBOX (for sync events) ----------------
type Outbox struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	EntityType string    `gorm:"index;not null"`
	EntityID   uuid.UUID `gorm:"type:uuid;not null"`
	Op         string    `gorm:"not null"` // UPSERT | DELETE | REINDEX_...
	Payload    datatypes.JSON
	CreatedAt  time.Time
	Processed  bool `gorm:"default:false"`
}
