package typescriptgen

import (
	"embed"

	"github.com/killean-solvely/hsapi-gen/internal/portal"
	"github.com/killean-solvely/hsapi-gen/pkg/templates"
)

//go:embed static/*.tstpl
var files embed.FS

type HubspotTSClientTemplateInput struct {
	PortalNames      map[string]string
	ObjectNameToType map[string]SchemaData
	AssociationTypes map[string]map[string]map[string]portal.Association
}

func GenerateTSClient(input HubspotTSClientTemplateInput) (string, error) {
	return templates.GenerateCode(files, "static/client.tstpl", input)
}

type TSPortalTemplateInput struct {
	PortalName       string
	AssociationTypes map[string]map[string]map[string]portal.Association
	Objects          map[string]string
}

func GenerateTSPortal(input TSPortalTemplateInput) (string, error) {
	return templates.GenerateCode(files, "static/portal.tstpl", input)
}

type TSSharedTemplateInput struct {
	AssociationTypes map[string]map[string]map[string]portal.Association
	Enums            []Enum
	Objects          []Object
}

func GenerateTSShared(input TSSharedTemplateInput) (string, error) {
	return templates.GenerateCode(files, "static/shared.tstpl", input)
}
