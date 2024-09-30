package codegen

import "time"

type SchemaData struct {
	InterfaceName string
	Description   string
	ObjectID      string
}

// SchemaResponse represents the response from HubSpot CRM's object schema API
type SchemaResponse struct {
	Results []Schema `json:"results"`
}

type Schema struct {
	Associations           []Association `json:"associations"`
	CreatedByUserID        int           `json:"createdByUserId"`
	ObjectTypeID           string        `json:"objectTypeId"`
	Description            string        `json:"description"`
	UpdatedByUserID        int           `json:"updatedByUserId"`
	FullyQualifiedName     string        `json:"fullyQualifiedName"`
	Labels                 Label         `json:"labels"`
	Archived               bool          `json:"archived"`
	CreatedAt              time.Time     `json:"createdAt"`
	PrimaryDisplayProperty string        `json:"primaryDisplayProperty"`
	Name                   string        `json:"name"`
	ID                     string        `json:"id"`
	Properties             []Property    `json:"properties"`
	UpdatedAt              time.Time     `json:"updatedAt"`
}

type Association struct {
	CreatedAt        time.Time `json:"createdAt"`        // When the association was defined
	FromObjectTypeID string    `json:"fromObjectTypeId"` // ID of the primary object type to link from
	Name             string    `json:"name"`             // A unique name for this association
	ID               string    `json:"id"`               // A unique ID for this association
	ToObjectTypeID   string    `json:"toObjectTypeId"`   // ID of the target object type to link to
	UpdatedAt        time.Time `json:"updatedAt"`        // When the association was last updated
}

type Label struct {
	Plural   string `json:"plural"`   // The word for multiple objects
	Singular string `json:"singular"` // The word for one object
}

type Property struct {
	Hidden               bool                 `json:"hidden"`
	DisplayOrder         int                  `json:"displayOrder"`
	Description          string               `json:"description"`
	ShowCurrencySymbol   bool                 `json:"showCurrencySymbol"`
	Type                 string               `json:"type"`
	HubspotDefined       bool                 `json:"hubspotDefined"`
	CreatedAt            time.Time            `json:"createdAt"`
	Archived             bool                 `json:"archived"`
	Options              []Option             `json:"options"`
	HasUniqueValue       bool                 `json:"hasUniqueValue"`
	Calculated           bool                 `json:"calculated"`
	ExternalOptions      bool                 `json:"externalOptions"`
	UpdatedAt            time.Time            `json:"updatedAt"`
	CreatedUserID        string               `json:"createdUserId"`
	ModificationMetadata ModificationMetadata `json:"modificationMetadata"`
	Label                string               `json:"label"`
	FormField            bool                 `json:"formField"`
	DataSensitivity      string               `json:"dataSensitivity"`
	ArchivedAt           time.Time            `json:"archivedAt"`
	GroupName            string               `json:"groupName"`
	ReferencedObjectType string               `json:"referencedObjectType"`
	Name                 string               `json:"name"`
	CalculationFormula   string               `json:"calculationFormula"`
	FieldType            string               `json:"fieldType"`
	UpdatedUserID        string               `json:"updatedUserId"`
}

type Option struct {
	Hidden       bool   `json:"hidden"`
	DisplayOrder int    `json:"displayOrder"`
	Description  string `json:"description"`
	Label        string `json:"label"`
	Value        string `json:"value"`
}

type ModificationMetadata struct {
	// fields for modification metadata
}
