package taskController

type createTaskPayload struct {
	Quantity    int    `json:"quantity"`
	Unit        string `json:"unit"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Frequency   string `json:"frequency"`
}
