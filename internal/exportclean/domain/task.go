package domain

type TaskFormat string
type TaskType string

const (
	FormatCSV  TaskFormat = "csv"
	FormatJSON TaskFormat = "json"

	TypeServers       TaskType = "servers"
	TypeHistory       TaskType = "history"
	TypeUserResources TaskType = "user_resources"
)

type ExportTask struct {
	ID         string
	ProjectKey string
	Format     TaskFormat
	Type       TaskType
	Limit      int
	Filters    map[string]string
}
