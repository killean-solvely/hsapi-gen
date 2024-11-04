package portal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/killean-solvely/hsapi-gen/internal/codegen/utils"
	"github.com/killean-solvely/hsapi-gen/pkg/hubspot"
)

type PortalDefinition struct {
	PortalName       string                                       `json:"portal_id"`
	Token            string                                       `json:"token"`
	Schemas          []hubspot.Schema                             `json:"schemas"`
	AssociationTypes map[string]map[string]map[string]Association `json:"association_types"`

	client   *hubspot.Client
	filename string
	logger   *log.Logger

	debug bool
}

func NewPortalDefinition(portalName, token string, debug bool) *PortalDefinition {
	filename := fmt.Sprintf("%s_api.json", portalName)
	logger := log.New(os.Stdout, "["+portalName+"] ", 0)

	return &PortalDefinition{
		PortalName:       portalName,
		Token:            token,
		Schemas:          []hubspot.Schema{},
		AssociationTypes: map[string]map[string]map[string]Association{},
		client:           hubspot.New(token),
		filename:         filename,
		logger:           logger,
		debug:            debug,
	}
}

func (pd *PortalDefinition) LoadPortalDefinition() error {
	// Check to see if the api file exists
	_, err := os.Stat(pd.filename)
	if err != nil {
		pd.logger.Println("Generating API file...")
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

		if pd.debug {
			pd.logger.Println("Saving API file...")
			err = pd.saveAPIToFile()
			if err != nil {
				return err
			}
		}
	} else {
		pd.logger.Println("Loading API file...")
		// If it does exist, load the api file
		err = pd.loadAPIFromFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func (pd PortalDefinition) getAllSchemas() ([]hubspot.Schema, error) {
	pd.logger.Println("Getting all schemas from portal...")

	schemas, err := pd.client.GetAllSchemas()
	if err != nil {
		return nil, err
	}

	pd.logger.Println("Default schemas retrieved.")

	return schemas, nil
}

func (pd PortalDefinition) getAssociationTypes(
	schemas []hubspot.Schema,
) (map[string]map[string]map[string]Association, error) {
	pd.logger.Println("Getting association types from HubSpot...")

	type LabelResponse struct {
		Results []Association `json:"results"`
	}

	// map of object to object to association label to association (associationTypes[fromObj][toObj][label] = association)
	associationTypes := map[string]map[string]map[string]Association{}
	for _, schema := range schemas {
		pd.logger.Printf(
			"Getting association types for schema %s, %d/%d\n",
			schema.Name,
			len(associationTypes)+1,
			len(schemas),
		)

		fromToLower := strings.ToLower(schema.Name)

		associationTypes[fromToLower] = map[string]map[string]Association{}
		for _, otherSchema := range schemas {
			if schema.Name == otherSchema.Name {
				continue
			}

			assocLabels, err := pd.client.GetAssociationLabels(
				schema.ObjectTypeID,
				otherSchema.ObjectTypeID,
			)
			if err != nil {
				pd.logger.Printf("Failed to get association labels: %s\n", err)
				return nil, err
			}

			if len(assocLabels) == 0 {
				continue
			}

			toToLower := strings.ToLower(otherSchema.Name)

			associationTypeName := fmt.Sprintf(
				"%s_to_%s",
				fromToLower,
				toToLower,
			)

			associationTypes[fromToLower][toToLower] = map[string]Association{}
			for _, assocLabel := range assocLabels {
				if assocLabel.Label == "" {
					sanitizedLabel := Association{
						AssociationLabel: assocLabel,
					}

					// If the label already exists, append a 2 to the end
					if _, ok := associationTypes[fromToLower][toToLower][associationTypeName]; ok {
						associationTypes[fromToLower][toToLower][associationTypeName+"2"] = sanitizedLabel
					} else {
						associationTypes[fromToLower][toToLower][associationTypeName] = sanitizedLabel
					}
				} else {
					sanitizedLabelName := utils.SanitizeLabel(assocLabel.Label)

					sanitizedLabel := Association{
						AssociationLabel: assocLabel,
						SanitizedLabel:   sanitizedLabelName,
					}

					// If the label already exists, append a 2 to the end
					if _, ok := associationTypes[fromToLower][toToLower][associationTypeName+"_"+sanitizedLabelName]; ok {
						associationTypes[fromToLower][toToLower][associationTypeName+"_"+sanitizedLabelName+"2"] = sanitizedLabel
					} else {
						associationTypes[fromToLower][toToLower][associationTypeName+"_"+sanitizedLabelName] = sanitizedLabel
					}
				}
			}
		}
	}

	pd.logger.Println("Association types retrieved.")

	return associationTypes, nil
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
		Schemas          []hubspot.Schema                             `json:"schemas"`
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
