package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type taskRaw struct {
	ID         string            `json:"id"`
	ProjectKey string            `json:"project_key"`
	Format     string            `json:"format"`
	Type       string            `json:"type"`
	Limit      int               `json:"limit"`
	Filters    map[string]string `json:"filters"`
}

type samples struct {
	Perfect   taskRaw `json:"perfect_sample"`
	Edge      taskRaw `json:"edge_sample"`
	Invalid   taskRaw `json:"invalid_sample"`
	TooLarge  taskRaw `json:"too_large_limit"`
	MissingID taskRaw `json:"missing_id"`
	TooLongPK taskRaw `json:"too_long_project"`
	TooLongFV taskRaw `json:"too_long_filter_value"`
	BadCharsP taskRaw `json:"invalid_char_project"`
	BadCharsF taskRaw `json:"invalid_char_filter"`
}

func loadSamplesAdapter(t *testing.T) samples {
	t.Helper()
	path := filepath.Join("..", "fixtures", "api_samples.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var s samples
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unmarshal fixtures: %v", err)
	}
	return s
}

func toDTO(raw taskRaw) *SubmitTaskDTO {
	return &SubmitTaskDTO{
		ID:         raw.ID,
		ProjectKey: raw.ProjectKey,
		Format:     raw.Format,
		Type:       raw.Type,
		Limit:      raw.Limit,
		Filters:    raw.Filters,
	}
}
