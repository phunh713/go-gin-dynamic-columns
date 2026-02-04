package constants

type Action string

const (
	ActionCreate  Action = "CREATE"
	ActionUpdate  Action = "UPDATE"
	ActionDelete  Action = "DELETE"
	ActionRefresh Action = "REFRESH"
)
