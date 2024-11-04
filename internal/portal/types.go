package portal

import "github.com/killean-solvely/hsapi-gen/pkg/hubspot"

type Association struct {
	hubspot.AssociationLabel
	SanitizedLabel string `json:"sanitized_label"`
}

type AssociationConfig struct {
	Associations map[string]map[string]map[string]Association `json:"associations"`
}
