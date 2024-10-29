package portal

type SchemaData struct {
	InterfaceName string
	Description   string
	ObjectID      string
}

type Association struct {
	ID             int    `json:"typeId"`
	Label          string `json:"label"`
	SanitizedLabel string `json:"sanitized_label"`
	Category       string `json:"category"`
}

type AssociationConfig struct {
	Associations map[string]map[string]map[string]Association `json:"associations"`
}

type Enum struct {
	Name   string
	Values map[string]string
}

type Property struct {
	Comment string
	Name    string
	Type    string
}

type Object struct {
	ID           string
	InternalName string
	Name         string
	Properties   []Property
}
