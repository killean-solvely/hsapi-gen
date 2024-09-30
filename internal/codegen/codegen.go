package codegen

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed snippets/*
var snippets embed.FS

type Codegen struct {
	hsToken string

	Schemas          []Schema
	AssociationTypes map[string]map[string]map[string]AssociationType
	ObjectNameToType map[string]SchemaData
	CodeSections     []string
}

func NewCodegen(hsToken string) *Codegen {
	return &Codegen{hsToken: hsToken}
}

func (c Codegen) GenerateAndSave(path string) error {
	err := c.getHubspotData()
	if err != nil {
		return err
	}

	err = c.generateCode()
	if err != nil {
		return err
	}

	err = c.SaveToFile(path)
	if err != nil {
		return err
	}

	return nil
}

// Gets the schemas and association types from hubspot
func (c *Codegen) getHubspotData() error {
	schemas, err := c.getAllSchemas()
	if err != nil {
		return err
	}
	c.Schemas = schemas

	associationTypes, err := c.getAssociationTypes(schemas)
	if err != nil {
		return err
	}
	c.AssociationTypes = associationTypes

	return nil
}

func (c *Codegen) generateCode() error {
	importAndConstantsContent, err := snippets.ReadFile("snippets/importsAndConstants.ts")
	if err != nil {
		return err
	}

	importsAndConstants := string(importAndConstantsContent)
	assocConfigCode := c.generateAssocConfigCode()
	c.ObjectNameToType = c.mapSchemaData()
	objectTypesCode := c.generateObjectTypesCode()
	typeToObjectIDCode := c.generateTypeToObjectIDListCode()
	hubspotClientCode, err := c.generateHubspotClientCode()
	if err != nil {
		return err
	}

	c.CodeSections = []string{
		importsAndConstants,
		assocConfigCode,
		objectTypesCode,
		typeToObjectIDCode,
		hubspotClientCode,
	}

	return nil
}

func (c Codegen) SaveToFile(path string) error {
	fullCode := strings.Join(c.CodeSections, "\n")

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(fullCode)
	if err != nil {
		return err
	}

	f.Close()

	return nil
}

func (c *Codegen) generateHubspotClientCode() (string, error) {
	functionBuildersContent, err := snippets.ReadFile("snippets/functionBuilders.ts")
	if err != nil {
		return "", err
	}
	functionBuilders := string(functionBuildersContent)

	sdkCode := "\n\tpublic api = {\n"
	for objectName, schemaData := range c.ObjectNameToType {
		if schemaData.Description != "" {
			sdkCode += fmt.Sprintf("\t\t/** %s */\n", schemaData.Description)
		}
		sdkCode += fmt.Sprintf("\t\t%s: {\n", objectName)
		sdkCode += fmt.Sprintf(
			"\t\t\tget: this.getObjectTypeFunction<\"%s\">(\"%s\"),\n",
			objectName,
			objectName,
		)
		sdkCode += fmt.Sprintf(
			"\t\t\tcreate: this.createObjectTypeFunction<\"%s\">(\"%s\"),\n",
			objectName,
			objectName,
		)
		sdkCode += fmt.Sprintf(
			"\t\t\tupdate: this.updateObjectTypeFunction<\"%s\">(\"%s\"),\n",
			objectName,
			objectName,
		)
		sdkCode += fmt.Sprintf(
			"\t\t\tgetAssociations: this.getAssociationsObjectTypeFunction<\"%s\">(\"%s\"),\n",
			objectName,
			objectName,
		)
		sdkCode += fmt.Sprintf(
			"\t\t\tassociate: this.associateObjectTypeFunction(\"%s\"),\n",
			objectName,
		)
		sdkCode += "\t\t},\n"
	}
	sdkCode += "\t}\n"
	sdkCode += "}\n"

	return functionBuilders + sdkCode, nil
}

func (c *Codegen) generateTypeToObjectIDListCode() string {
	typeToObjectIDCode := "const TypeToObjectIDList = {\n"
	for objectName, schemaData := range c.ObjectNameToType {
		typeToObjectIDCode += fmt.Sprintf("\t%s: \"%s\",\n", objectName, schemaData.ObjectID)
	}
	typeToObjectIDCode += "} as const;\n\n"
	typeToObjectIDCode += "type TypeToObjectIDList = typeof TypeToObjectIDList;\n"
	typeToObjectIDCode += "type TypeKeys = keyof TypeToObjectIDList;\n"
	return typeToObjectIDCode
}

func (c *Codegen) generateObjectTypesCode() string {
	schemaCodeSections := []string{}
	for _, schema := range c.Schemas {
		schemaEnums := ""
		createdEnums := map[string]bool{}

		lowerSchemaName := strings.ToLower(schema.Name)

		schemaCode := fmt.Sprintf(
			"interface %s {\n",
			c.ObjectNameToType[lowerSchemaName].InterfaceName,
		)
		for _, prop := range schema.Properties {
			propertyType := prop.Type
			propertyLabel := prop.Label
			propertyName := prop.Name

			possibleEnumName := fmt.Sprintf(
				"%s%s",
				c.ObjectNameToType[lowerSchemaName].InterfaceName,
				convertLabelToEnumName(propertyLabel),
			)

			isEnumeration := propertyType == "enumeration"
			isEmptyEnumeration := len(prop.Options) == 0

			propType, ok := typeConversionMap[propertyType]
			if !ok {
				propType = propertyType
			}

			if isEnumeration && !isEmptyEnumeration {
				propType = possibleEnumName + "Enum"
			} else if isEmptyEnumeration {
				propType = "string"
			}

			if prop.Description != "" {
				schemaCode += fmt.Sprintf("\t/** %s */\n", prop.Description)
			}
			schemaCode += fmt.Sprintf("\t%s: %s;\n", propertyName, propType)

			_, enumExists := createdEnums[possibleEnumName]
			if isEnumeration && !enumExists && !(isEnumeration && isEmptyEnumeration) {
				createdEnums[possibleEnumName] = true

				schemaEnums += fmt.Sprintf("export enum %sEnum {\n", possibleEnumName)
				for _, option := range prop.Options {
					sanitizedOptionLabel := sanitizeLabel(option.Label)
					if sanitizedOptionLabel == "" {
						sanitizedOptionLabel = "_"
					}
					schemaEnums += fmt.Sprintf(
						"\t%s = \"%s\",\n",
						prependUnderscoreToEnum(sanitizedOptionLabel),
						strings.ReplaceAll(option.Value, "\"", "\\\""),
					)
				}
				schemaEnums += "}\n\n"
			}
		}

		schemaCode += "}\n"

		finalCode := schemaEnums + schemaCode
		schemaCodeSections = append(schemaCodeSections, finalCode)
	}

	objectTypesCode := "interface ObjectTypes {\n"
	for objectName, schemaData := range c.ObjectNameToType {
		objectTypesCode += fmt.Sprintf("\t%s: %s;\n", objectName, schemaData.InterfaceName)
	}
	objectTypesCode += "}\n"

	schemaCodeSections = append(schemaCodeSections, objectTypesCode)

	return strings.Join(schemaCodeSections, "\n")
}

func (c *Codegen) mapSchemaData() map[string]SchemaData {
	objectNameToType := map[string]SchemaData{}
	for _, schema := range c.Schemas {
		objectNameToType[strings.ToLower(schema.Name)] = SchemaData{
			InterfaceName: fmt.Sprintf("%sProps", convertSchemaNameToInterfaceName(schema.Name)),
			Description:   schema.Description,
			ObjectID:      schema.ObjectTypeID,
		}
	}
	return objectNameToType
}

func (c *Codegen) generateAssocConfigCode() string {
	assocTypesCode := "const AssociationsConfig = {\n"
	for objName, otherObjAssocMap := range c.AssociationTypes {
		assocTypesCode += fmt.Sprintf("\t%s: {\n", objName)
		for otherObjName, assocOptions := range otherObjAssocMap {
			if len(assocOptions) == 0 {
				continue
			}

			assocTypesCode += fmt.Sprintf("\t\t%s: {\n", otherObjName)
			for assocTypeStr, assocData := range assocOptions {
				assocTypesCode += fmt.Sprintf("\t\t\t%s: {\n", assocTypeStr)
				assocTypesCode += fmt.Sprintf("\t\t\t\tID: %d,\n", assocData.ID)
				assocTypesCode += fmt.Sprintf(
					"\t\t\t\tCategory: AssociationSpecAssociationCategoryEnum.%s,\n",
					categoryMap[assocData.Category],
				)
				assocTypesCode += "\t\t\t},\n"
			}
			assocTypesCode += "\t\t},\n"
		}
		assocTypesCode += "\t},\n"
	}
	assocTypesCode += "} as const;\n\n"
	assocTypesCode += "type AssociationsConfig = typeof AssociationsConfig;\n"
	assocTypesCode += "type AssociationKeys<F extends keyof AssociationsConfig, T extends keyof AssociationsConfig[F]> = keyof AssociationsConfig[F][T];"
	return assocTypesCode
}
