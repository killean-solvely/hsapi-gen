package templates

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/killean-solvely/hsapi-gen/internal/codegen/portal"
)

//go:embed static/*.tstpl
var templates embed.FS

func createTemplate(name, t string) *template.Template {
	return template.Must(template.New(name).Parse(t))
}

func generateCode[T any](input T, templatePath string) (string, error) {
	f, err := templates.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	t := createTemplate("temp", string(f))

	var buf bytes.Buffer
	err = t.Execute(&buf, input)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type HubspotClientTemplateInput struct {
	PortalNames      map[string]string
	ObjectNameToType map[string]portal.SchemaData
	AssociationTypes map[string]map[string]map[string]portal.Association
}

func GenerateClient(input HubspotClientTemplateInput) (string, error) {
	return generateCode(input, "static/client.tstpl")
}

type PortalTemplateInput struct {
	PortalName       string
	AssociationTypes map[string]map[string]map[string]portal.Association
	Objects          map[string]string
}

func GeneratePortal(input PortalTemplateInput) (string, error) {
	return generateCode(input, "static/portal.tstpl")
}

type SharedTemplateInput struct {
	AssociationTypes map[string]map[string]map[string]portal.Association
	Enums            []portal.Enum
	Objects          []portal.Object
}

func GenerateShared(input SharedTemplateInput) (string, error) {
	return generateCode(input, "static/shared.tstpl")
}
