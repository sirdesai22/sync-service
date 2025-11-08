// internal/elastic/index.go
package elastic

import (
	"bytes"
	"context"
	"fmt"
	es "github.com/elastic/go-elasticsearch/v8"
)

const (
	IdxUsers     = "users_v1"
	IdxHackathons= "hackathons_v1"
	IdxProjects  = "projects_v1"
)

func EnsureIndexes(ctx context.Context, c *es.Client) error {
	mapping := `{"settings":{"number_of_shards":1},"mappings":{"dynamic":"strict","properties":{
		"username":{"type":"keyword"},"email":{"type":"keyword"},"skills":{"type":"keyword"},
		"college":{"type":"text"},"updated_at":{"type":"date"}
	}}}`
	if err := ensure(ctx, c, IdxUsers, mapping); err != nil { return err }

	mapping = `{"settings":{"number_of_shards":1},"mappings":{"dynamic":"strict","properties":{
		"name":{"type":"text"},"location":{"type":"keyword"},"tracks":{"type":"keyword"},
		"start_at":{"type":"date"},"end_at":{"type":"date"},"updated_at":{"type":"date"}
	}}}`
	if err := ensure(ctx, c, IdxHackathons, mapping); err != nil { return err }

	mapping = `{"settings":{"number_of_shards":1},"mappings":{"dynamic":"strict","properties":{
		"name":{"type":"text"},"description":{"type":"text"},"hackathon_id":{"type":"keyword"},
		"owner_id":{"type":"keyword"},"team_members":{"type":"keyword"},"updated_at":{"type":"date"}
	}}}`
	return ensure(ctx, c, IdxProjects, mapping)
}

func ensure(ctx context.Context, c *es.Client, index, body string) error {
	exists, _ := c.Indices.Exists([]string{index})
	if exists.StatusCode == 200 { return nil }
	_, err := c.Indices.Create(index, c.Indices.Create.WithBody(bytes.NewBufferString(body)), c.Indices.Create.WithContext(ctx))
	if err != nil { return fmt.Errorf("create index %s: %w", index, err) }
	return nil
}
