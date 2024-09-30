package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"
)

//go:embed snippets/*
var snippets embed.FS

type Codegen struct {
	hsToken string
}

func NewCodegen(hsToken string) *Codegen {
	return &Codegen{hsToken: hsToken}
}

// sanitizeLabel removes special characters, replaces spaces with underscores, and converts to lowercase.
func sanitizeLabel(input string) string {
	if input == "" {
		return "_"
	}
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ToLower(input)
	return input
}

// convertSchemaNameToInterfaceName removes special characters, converts to lowercase, capitalizes first letter of each word, and joins them.
func convertSchemaNameToInterfaceName(input string) string {
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ToLower(input)

	words := strings.Split(input, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	return strings.Join(words, "")
}

// prependUnderscoreToEnum adds underscore to the beginning if the string starts with a digit.
func prependUnderscoreToEnum(e string) string {
	reg, _ := regexp.Compile(`^(\d)`)
	return reg.ReplaceAllString(e, "_$1")
}

// convertLabelToEnumName removes special characters, converts to lowercase, capitalizes first letter of each word, and joins them.
func convertLabelToEnumName(input string) string {
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ToLower(input)

	words := strings.Fields(input)
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	return strings.Join(words, "")
}

// strings.Title is not directly analogous to TypeScript's `charAt(0).toUpperCase()` as it capitalizes each
// word in the string. We define a proper toFirstUpper function.
func toFirstUpper(word string) string {
	for i, v := range word {
		return string(unicode.ToUpper(v)) + word[i+1:]
	}
	return ""
}

func (c Codegen) getAllSchemas() ([]Schema, error) {
	schemas := []Schema{}

	req, err := http.NewRequest("GET", "https://api.hubapi.com/crm-object-schemas/v3/schemas", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %s\n", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.hsToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err)
		return nil, err
	}

	var schemasResponse SchemaResponse
	err = json.Unmarshal(body, &schemasResponse)
	if err != nil {
		fmt.Printf("Failed to unmarshal response: %s\n", err)
		return nil, err
	}

	for i := range schemasResponse.Results {
		schemas = append(schemas, schemasResponse.Results[i])
	}

	objectTypes := []string{
		"call",
		"cart",
		"communication",
		"company",
		"contact",
		"deal",
		"discount",
		"email",
		"engagement",
		"fee",
		"feedback_submission",
		"goal_target",
		"line_item",
		"marketing_event",
		"meeting_event",
		"note",
		"order",
		"postal_mail",
		"product",
		"quote",
		"quote_template",
		"task",
		"tax",
		"ticket",
	}

	for _, objectType := range objectTypes {
		req, err := http.NewRequest(
			"GET",
			"https://api.hubapi.com/crm-object-schemas/v3/schemas/"+objectType,
			nil,
		)
		if err != nil {
			fmt.Printf("Failed to create request: %s\n", err)
			return nil, err
		}
		// Replace YOUR_ACCESS_TOKEN with the actual token
		req.Header.Add("Authorization", "Bearer "+c.hsToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Failed to send request: %s\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %s\n", err)
			return nil, err
		}

		var schema Schema
		err = json.Unmarshal(body, &schema)
		if err != nil {
			fmt.Printf("Failed to unmarshal response: %s\n", err)
			return nil, err
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

type AssociationType struct {
	ID       int    `json:"typeId"`
	Label    string `json:"label"`
	Category string `json:"category"` // hubspot defined or user defined
}

func (c Codegen) getAssociationTypes(
	schemas []Schema,
) (map[string]map[string]map[string]AssociationType, error) {
	type LabelResponse struct {
		Results []AssociationType `json:"results"`
	}

	// map of association string type to association type
	associationTypes := map[string]map[string]map[string]AssociationType{}
	for _, schema := range schemas {
		associationTypes[strings.ToLower(schema.Name)] = map[string]map[string]AssociationType{}

		for _, otherSchema := range schemas {
			if schema.Name == otherSchema.Name {
				continue
			}

			associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)] = map[string]AssociationType{}

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf(
					"https://api.hubapi.com/crm/v4/associations/%s/%s/labels",
					schema.ObjectTypeID,
					otherSchema.ObjectTypeID,
				),
				nil,
			)
			if err != nil {
				fmt.Printf("Failed to create request: %s\n", err)
				return nil, err
			}
			req.Header.Add("Authorization", "Bearer "+c.hsToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("Failed to send request: %s\n", err)
				return nil, err
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read response body: %s\n", err)
				return nil, err
			}

			var labelResponse LabelResponse
			err = json.Unmarshal(body, &labelResponse)
			if err != nil {
				fmt.Printf("Failed to unmarshal response: %s\n", err)
				return nil, err
			}

			if len(labelResponse.Results) == 0 {
				continue
			}

			associationTypeName := fmt.Sprintf(
				"%s_to_%s",
				strings.ToLower(schema.Name),
				strings.ToLower(otherSchema.Name),
			)

			for _, associationType := range labelResponse.Results {
				if associationType.Label == "" {
					if _, ok := associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName]; ok {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"2"] = associationType
					} else {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName] = associationType
					}
				} else {
					assocLabel := sanitizeLabel(associationType.Label)
					if _, ok := associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel]; ok {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel+"2"] = associationType
					} else {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel] = associationType
					}
				}
			}
		}
	}

	return associationTypes, nil
}

func (c *Codegen) Generate() error {
	TypeConversionMap := map[string]string{
		"date":               "string",
		"datetime":           "string",
		"bool":               "boolean",
		"object_coordinates": "string",
		"json":               "string",
		"phone_number":       "string",
	}

	importAndConstantsContent, err := snippets.ReadFile("snippets/importsAndConstants.ts")
	if err != nil {
		return err
	}
	importsAndConstants := string(importAndConstantsContent)

	schemas, err := c.getAllSchemas()
	if err != nil {
		return err
	}

	associationTypes, err := c.getAssociationTypes(schemas)
	if err != nil {
		return err
	}

	categoryMap := map[string]string{
		"HUBSPOT_DEFINED":    "HubspotDefined",
		"USER_DEFINED":       "UserDefined",
		"INTEGRATOR_DEFINED": "IntegratorDefined",
	}

	assocTypesCode := "const AssociationsConfig = {\n"
	for objName, otherObjAssocMap := range associationTypes {
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

	type SchemaData struct {
		InterfaceName string
		Description   string
		ObjectID      string
	}

	objectNameToType := map[string]SchemaData{}

	schemaCodeSections := []string{importsAndConstants, assocTypesCode}
	for _, schema := range schemas {
		schemaCode := ""

		schemaInterfaceName := fmt.Sprintf("%sProps", convertSchemaNameToInterfaceName(schema.Name))
		objectNameToType[strings.ToLower(schema.Name)] = SchemaData{
			InterfaceName: schemaInterfaceName,
			Description:   schema.Description,
			ObjectID:      schema.ObjectTypeID,
		}

		schemaEnums := ""
		createdEnums := map[string]bool{}

		schemaCode += fmt.Sprintf("interface %s {\n", schemaInterfaceName)
		for _, prop := range schema.Properties {
			propertyType := prop.Type
			propertyLabel := prop.Label
			propertyName := prop.Name

			possibleEnumName := fmt.Sprintf(
				"%s%s",
				schemaInterfaceName,
				convertLabelToEnumName(propertyLabel),
			)

			isEnumeration := propertyType == "enumeration"
			isEmptyEnumeration := len(prop.Options) == 0

			propType, ok := TypeConversionMap[propertyType]
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

	typeToObjectIDCode := "const TypeToObjectIDList = {\n"
	for objectName, schemaData := range objectNameToType {
		typeToObjectIDCode += fmt.Sprintf("\t%s: \"%s\",\n", objectName, schemaData.ObjectID)
	}
	typeToObjectIDCode += "} as const;\n\n"
	typeToObjectIDCode += "type TypeToObjectIDList = typeof TypeToObjectIDList;\n"
	typeToObjectIDCode += "type TypeKeys = keyof TypeToObjectIDList;\n"

	schemaCodeSections = append(schemaCodeSections, typeToObjectIDCode)

	objectTypesCode := "interface ObjectTypes {\n"
	for objectName, schemaData := range objectNameToType {
		objectTypesCode += fmt.Sprintf("\t%s: %s;\n", objectName, schemaData.InterfaceName)
	}
	objectTypesCode += "}\n"

	schemaCodeSections = append(schemaCodeSections, objectTypesCode)

	functionBuildersContent, err := snippets.ReadFile("snippets/functionBuilders.ts")
	if err != nil {
		return err
	}
	functionBuilders := string(functionBuildersContent)

	schemaCodeSections = append(schemaCodeSections, functionBuilders)

	sdkCode := "\tpublic api = {\n"
	for objectName, schemaData := range objectNameToType {
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

	schemaCodeSections = append(schemaCodeSections, sdkCode)

	fullCode := strings.Join(schemaCodeSections, "\n")

	f, err := os.Create("generated.ts")
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
