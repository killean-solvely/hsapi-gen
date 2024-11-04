package typescriptgen

import (
	"fmt"
	"strings"

	"github.com/killean-solvely/hsapi-gen/internal/codegen/utils"
	"github.com/killean-solvely/hsapi-gen/internal/portal"
)

type SharedDefinition struct {
	portal         *portal.PortalDefinition
	codeDefinition *PortalCodeDefinition
}

type SchemaData struct {
	ObjectID      string
	InterfaceName string
	Description   string
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
	ID            string
	InternalName  string
	InterfaceName string
	Properties    []Property
}

type PortalCodeDefinition struct {
	OriginalPortal *portal.PortalDefinition

	ObjectNameToType map[string]SchemaData `json:"object_name_to_type"`
	Enums            []Enum                `json:"enums"`
	Objects          []Object              `json:"objects"`
	ObjectIDs        map[string]string     `json:"object_ids"`
}

func (pcd *PortalCodeDefinition) parsePortalDefinition() {
	pcd.parseSchemaData()
}

func (pcd *PortalCodeDefinition) parseSchemaData() {
	pcd.ObjectNameToType = map[string]SchemaData{}
	for _, schema := range pcd.OriginalPortal.Schemas {
		pcd.ObjectNameToType[strings.ToLower(schema.Name)] = SchemaData{
			ObjectID:      schema.ObjectTypeID,
			InterfaceName: utils.ConvertSchemaNameToInterfaceName(schema.Name),
			Description:   schema.Description,
		}
	}
}

func (pcd *PortalCodeDefinition) parseObjects() {
	createdEnums := map[string]bool{}

	for _, schema := range pcd.OriginalPortal.Schemas {
		lowerSchemaName := strings.ToLower(schema.Name)
		obj := Object{}
		obj.ID = schema.ObjectTypeID
		obj.InternalName = lowerSchemaName
		obj.InterfaceName = pcd.ObjectNameToType[lowerSchemaName].InterfaceName

		for _, prop := range schema.Properties {
			possibleEnumName := fmt.Sprintf(
				"%s%s",
				pcd.ObjectNameToType[lowerSchemaName].InterfaceName,
				utils.ConvertLabelToEnumName(prop.Label),
			)

			isEnumeration := prop.Type == "enumeration"
			isEmptyEnumeration := len(prop.Options) == 0

			propType, ok := typeConversionMap[prop.Type]
			if !ok {
				propType = prop.Type
			}

			_, enumExists := createdEnums[possibleEnumName]
			if isEnumeration && !enumExists && !(isEnumeration && isEmptyEnumeration) {
				createdEnums[possibleEnumName] = true

				propType = possibleEnumName + "Enum"
				enumOptions := map[string]string{}
				for _, option := range prop.Options {
					sanitizedOptionLabel := utils.SanitizeLabel(option.Label)
					if sanitizedOptionLabel == "" {
						sanitizedOptionLabel = "_"
					}
					enumOptions[utils.PrependUnderscoreToEnum(sanitizedOptionLabel)] = strings.ReplaceAll(
						option.Value,
						"\"",
						"\\\"",
					)
				}

				pcd.Enums = append(pcd.Enums, Enum{
					Name:   propType,
					Values: enumOptions,
				})
			} else if isEmptyEnumeration {
				propType = "string"
			} else {
				propType = "string"
			}

			obj.Properties = append(obj.Properties, Property{
				Comment: prop.Description,
				Name:    prop.Name,
				Type:    propType,
			})
		}

		pcd.Objects = append(pcd.Objects, obj)
		pcd.ObjectIDs[pcd.ObjectNameToType[lowerSchemaName].InterfaceName] = schema.ObjectTypeID
	}
}
