package codegen

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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
