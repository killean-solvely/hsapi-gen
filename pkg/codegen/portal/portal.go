package portal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/killean-solvely/hsapi-gen/pkg/codegen/hs"
	"github.com/killean-solvely/hsapi-gen/pkg/codegen/utils"
)

type PortalDefinition struct {
	PortalName       string                                       `json:"portal_id"`
	Token            string                                       `json:"token"`
	Schemas          []hs.Schema                                  `json:"schemas"`
	AssociationTypes map[string]map[string]map[string]Association `json:"association_types"`
	ObjectNameToType map[string]SchemaData                        `json:"object_name_to_type"`
	filename         string
	logger           *log.Logger

	Enums     []Enum            `json:"enums"`
	Objects   []Object          `json:"objects"`
	ObjectIDs map[string]string `json:"object_ids"`
}

func NewPortalDefinition(portalName, token string, logger *log.Logger) *PortalDefinition {
	filename := fmt.Sprintf("%s_api.json", portalName)
	objectIDs := map[string]string{}
	//logger := log.New(logger.Panicf, "["+portalName+"] ", 0)

	return &PortalDefinition{
		PortalName:       portalName,
		Token:            token,
		Schemas:          []hs.Schema{},
		AssociationTypes: map[string]map[string]map[string]Association{},
		ObjectNameToType: map[string]SchemaData{},
		filename:         filename,
		logger:           logger,
		Enums:            []Enum{},
		Objects:          []Object{},
		ObjectIDs:        objectIDs,
	}
}

func (pd *PortalDefinition) LoadPortalDefinition() error {
	// Check to see if the api file exists
	_, err := os.Stat(pd.filename)
	if err != nil {
		pd.logger.Println("[" + pd.PortalName + "] " + "Generating API file...")
		// If it doesn't exist, generate the api file
		schemas, err := pd.getAllSchemas()
		if err != nil {
			return err
		}
		pd.Schemas = schemas

		associationTypes, err := pd.getAssociationTypes(schemas)
		if err != nil {
			return err
		}
		pd.AssociationTypes = associationTypes

		// Disabled in production since people probably won't need this
		// pd.logger.Println("Saving API file...")
		// err = pd.saveAPIToFile()
		// if err != nil {
		// 	return err
		// }
	} else {
		pd.logger.Println("[" + pd.PortalName + "] " + "Loading API file...")
		// If it does exist, load the api file
		err = pd.loadAPIFromFile()
		if err != nil {
			return err
		}
	}

	pd.logger.Println("[" + pd.PortalName + "] " + "Parsing API data...")
	pd.parseData()
	pd.logger.Println("[" + pd.PortalName + "] " + "API data parsed.")

	return nil
}

func (pd PortalDefinition) getAllSchemas() ([]hs.Schema, error) {
	pd.logger.Println("[" + pd.PortalName + "] " + "Getting all schemas from portal...")

	schemas := []hs.Schema{}

	req, err := http.NewRequest("GET", "https://api.hubapi.com/crm-object-schemas/v3/schemas", nil)
	if err != nil {
		pd.logger.Printf("["+pd.PortalName+"] "+"Failed to create request: %s\n", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+pd.Token)

	pd.logger.Println("[" + pd.PortalName + "] " + "Getting custom schemas...")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		pd.logger.Printf("["+pd.PortalName+"] "+"Failed to send request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pd.logger.Printf("["+pd.PortalName+"] "+"Failed to read response body: %s\n", err)
		return nil, err
	}

	var schemasResponse hs.SchemaResponse
	err = json.Unmarshal(body, &schemasResponse)
	if err != nil {
		pd.logger.Printf("["+pd.PortalName+"] "+"Failed to unmarshal response: %s\n", err)
		return nil, err
	}

	for i := range schemasResponse.Results {
		schemas = append(schemas, schemasResponse.Results[i])
	}

	pd.logger.Println("[" + pd.PortalName + "] " + "Custom schemas retrieved.")

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

	pd.logger.Printf("["+pd.PortalName+"] "+"Getting %d default schemas...\n", len(objectTypes))

	for i, objectType := range objectTypes {
		pd.logger.Printf(
			"["+pd.PortalName+"] "+"Getting schema %d/%d: %s\n",
			i+1,
			len(objectTypes),
			objectType,
		)

		req, err := http.NewRequest(
			"GET",
			"https://api.hubapi.com/crm-object-schemas/v3/schemas/"+objectType,
			nil,
		)
		if err != nil {
			pd.logger.Printf("["+pd.PortalName+"] "+"Failed to create request: %s\n", err)
			return nil, err
		}
		// Replace YOUR_ACCESS_TOKEN with the actual token
		req.Header.Add("Authorization", "Bearer "+pd.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			pd.logger.Printf("["+pd.PortalName+"] "+"Failed to send request: %s\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pd.logger.Printf("["+pd.PortalName+"] "+"Failed to read response body: %s\n", err)
			return nil, err
		}

		var schema hs.Schema
		err = json.Unmarshal(body, &schema)
		if err != nil {
			pd.logger.Printf("["+pd.PortalName+"] "+"Failed to unmarshal response: %s\n", err)
			return nil, err
		}

		schemas = append(schemas, schema)
	}

	pd.logger.Println("[" + pd.PortalName + "] " + "Default schemas retrieved.")

	return schemas, nil
}

func (pd PortalDefinition) getAssociationTypes(
	schemas []hs.Schema,
) (map[string]map[string]map[string]Association, error) {
	pd.logger.Println("[" + pd.PortalName + "] " + "Getting association types from HubSpot...")

	type LabelResponse struct {
		Results []Association `json:"results"`
	}

	// map of association string type to association type
	associationTypes := map[string]map[string]map[string]Association{}
	for _, schema := range schemas {
		pd.logger.Printf(
			"["+pd.PortalName+"] "+
				"Getting association types for schema %s, %d/%d\n",
			schema.Name,
			len(associationTypes)+1,
			len(schemas),
		)

		associationTypes[strings.ToLower(schema.Name)] = map[string]map[string]Association{}

		for _, otherSchema := range schemas {
			if schema.Name == otherSchema.Name {
				continue
			}

			associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)] = map[string]Association{}

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
				pd.logger.Printf("["+pd.PortalName+"] "+"Failed to create request: %s\n", err)
				return nil, err
			}
			req.Header.Add("Authorization", "Bearer "+pd.Token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				pd.logger.Printf("["+pd.PortalName+"] "+"Failed to send request: %s\n", err)
				return nil, err
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				pd.logger.Printf("["+pd.PortalName+"] "+"Failed to read response body: %s\n", err)
				return nil, err
			}

			var labelResponse LabelResponse
			err = json.Unmarshal(body, &labelResponse)
			if err != nil {
				pd.logger.Printf("["+pd.PortalName+"] "+"Failed to unmarshal response: %s\n", err)
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
					assocLabel := utils.SanitizeLabel(associationType.Label)
					associationType.SanitizedLabel = assocLabel

					if _, ok := associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel]; ok {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel+"2"] = associationType
					} else {
						associationTypes[strings.ToLower(schema.Name)][strings.ToLower(otherSchema.Name)][associationTypeName+"_"+assocLabel] = associationType
					}
				}
			}
		}
	}

	pd.logger.Println("[" + pd.PortalName + "] " + "Association types retrieved.")

	return associationTypes, nil
}

func (pd *PortalDefinition) parseData() {
	pd.parseSchemaData()
	pd.parseObjects()
}

func (pd *PortalDefinition) parseSchemaData() {
	pd.ObjectNameToType = map[string]SchemaData{}
	for _, schema := range pd.Schemas {
		pd.ObjectNameToType[strings.ToLower(schema.Name)] = SchemaData{
			InterfaceName: fmt.Sprintf(
				"%s",
				utils.ConvertSchemaNameToInterfaceName(schema.Name),
			),
			Description: schema.Description,
			ObjectID:    schema.ObjectTypeID,
		}
	}
}

func (pd *PortalDefinition) parseObjects() {
	createdEnums := map[string]bool{}

	for _, schema := range pd.Schemas {
		lowerSchemaName := strings.ToLower(schema.Name)
		obj := Object{}
		obj.ID = schema.ObjectTypeID
		obj.InternalName = lowerSchemaName
		obj.Name = pd.ObjectNameToType[lowerSchemaName].InterfaceName

		for _, prop := range schema.Properties {
			propertyType := prop.Type
			propertyLabel := prop.Label
			propertyName := prop.Name

			possibleEnumName := fmt.Sprintf(
				"%s%s",
				pd.ObjectNameToType[lowerSchemaName].InterfaceName,
				utils.ConvertLabelToEnumName(propertyLabel),
			)

			isEnumeration := propertyType == "enumeration"
			isEmptyEnumeration := len(prop.Options) == 0

			propType, ok := typeConversionMap[propertyType]
			if !ok {
				propType = propertyType
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

				pd.Enums = append(pd.Enums, Enum{
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
				Name:    propertyName,
				Type:    propType,
			})
		}

		pd.Objects = append(pd.Objects, obj)
		pd.ObjectIDs[pd.ObjectNameToType[lowerSchemaName].InterfaceName] = schema.ObjectTypeID
	}
}

func (pd PortalDefinition) saveAPIToFile() error {
	data, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	err = os.WriteFile(pd.filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (pd *PortalDefinition) loadAPIFromFile() error {
	data, err := os.ReadFile(pd.filename)
	if err != nil {
		return err
	}

	var apiFile struct {
		Schemas          []hs.Schema                                  `json:"schemas"`
		AssociationTypes map[string]map[string]map[string]Association `json:"association_types"`
	}
	err = json.Unmarshal(data, &apiFile)
	if err != nil {
		return err
	}

	pd.Schemas = apiFile.Schemas
	pd.AssociationTypes = apiFile.AssociationTypes

	return nil
}
