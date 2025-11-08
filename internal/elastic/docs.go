// internal/elastic/docs.go
package elastic

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
	"github.com/sirdesai22/sync-service/internal/models"
)

type UserDoc struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Skills    []string  `json:"skills"`
	College   string    `json:"college"`
	UpdatedAt time.Time `json:"updated_at"`
}
func BuildUserDoc(u models.User) ([]byte, error) {
	var skills []string
	_ = json.Unmarshal(u.Skills, &skills)
	return json.Marshal(UserDoc{
		Username: u.Username, Email: u.Email, Skills: skills, College: u.College, UpdatedAt: u.UpdatedAt,
	})
}

type HackathonDoc struct {
	Name string `json:"name"`; Location string `json:"location"`
	Tracks []string `json:"tracks"`; StartAt time.Time `json:"start_at"`; EndAt time.Time `json:"end_at"`; UpdatedAt time.Time `json:"updated_at"`
}
func BuildHackathonDoc(h models.Hackathon) ([]byte, error) {
	var tracks []string; _ = json.Unmarshal(h.Tracks, &tracks)
	return json.Marshal(HackathonDoc{h.Name, h.Location, tracks, h.StartAt, h.EndAt, h.UpdatedAt})
}

type ProjectDoc struct {
	Name string `json:"name"`; Description string `json:"description"`
	HackathonID uuid.UUID `json:"hackathon_id"`; OwnerID uuid.UUID `json:"owner_id"`
	TeamMembers []string `json:"team_members"`; UpdatedAt time.Time `json:"updated_at"`
}
func BuildProjectDoc(p models.Project) ([]byte, error) {
	var members []string; _ = json.Unmarshal(p.TeamMembers, &members)
	return json.Marshal(ProjectDoc{p.Name, p.Description, p.HackathonID, p.OwnerID, members, p.UpdatedAt})
}
