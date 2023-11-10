package task_management_workitem_schema

const (
	TEXT       = "text"
	LONGTEXT   = "longtext"
	URL        = "url"
	IDENTIFIER = "identifier"
	ENUM       = "enum"
	DATE       = "date"
	DATETIME   = "datetime"
	INTEGER    = "integer"
	NUMBER     = "number"
	CHECKBOX   = "checkbox"
	TAG        = "tag"
)

type textTypeField struct {
	title       string
	description string
	varType     string
	minLength   int
	maxLength   int
}

type numTypeField struct {
	title       string
	description string
	varType     string
	minVal      int
	maxVal      int
}
