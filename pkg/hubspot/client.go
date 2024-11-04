package hubspot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type customTransport struct {
	apikey string
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+c.apikey)
	return http.DefaultTransport.RoundTrip(clonedReq)
}

type Client struct {
	client *http.Client
	logger *log.Logger
}

func New(apikey string) *Client {
	client := &http.Client{
		Transport: &customTransport{
			apikey: apikey,
		},
	}

	logger := log.New(os.Stdout, "[hs_api] ", 0)

	return &Client{
		client: client,
		logger: logger,
	}
}

func (c Client) GetCustomSchemas() ([]Schema, error) {
	req, err := http.NewRequest("GET", "https://api.hubapi.com/crm-object-schemas/v3/schemas", nil)
	if err != nil {
		c.logger.Printf("Failed to create request: %s\n", err)
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Printf("Failed to do request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Failed to read response body: %s\n", err)
		return nil, err
	}

	var schemasResponse SchemaResponse
	err = json.Unmarshal(body, &schemasResponse)
	if err != nil {
		c.logger.Printf("Failed to unmarshal response body: %s\n", err)
		return nil, err
	}

	return schemasResponse.Results, nil
}

func (c Client) GetDefaultSchemas() ([]Schema, error) {
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

	schemas := []Schema{}
	for _, objectType := range objectTypes {
		req, err := http.NewRequest(
			"GET",
			"https://api.hubapi.com/crm-object-schemas/v3/schemas/"+objectType,
			nil,
		)
		if err != nil {
			c.logger.Printf("Failed to create request: %s\n", err)
			return nil, err
		}

		resp, err := c.client.Do(req)
		if err != nil {
			c.logger.Printf("Failed to do request: %s\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Printf("Failed to read response body: %s\n", err)
			return nil, err
		}

		var schema Schema
		err = json.Unmarshal(body, &schema)
		if err != nil {
			c.logger.Printf("Failed to unmarshal response body: %s\n", err)
			return nil, err
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (c Client) GetAllSchemas() ([]Schema, error) {
	customSchemas, err := c.GetCustomSchemas()
	if err != nil {
		return nil, err
	}

	defaultSchemas, err := c.GetDefaultSchemas()
	if err != nil {
		return nil, err
	}

	return append(customSchemas, defaultSchemas...), nil
}

func (c Client) GetAssociationLabels(
	fromObjTypeID string,
	toObjTypeID string,
) ([]AssociationLabel, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"https://api.hubapi.com/crm/v4/associations/%s/%s/labels",
			fromObjTypeID,
			toObjTypeID,
		),
		nil,
	)
	if err != nil {
		c.logger.Printf("Failed to create request: %s\n", err)
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Printf("Failed to do request: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Failed to read response body: %s\n", err)
		return nil, err
	}

	var labelResponse AssociationLabelResponse
	err = json.Unmarshal(body, &labelResponse)
	if err != nil {
		c.logger.Printf("Failed to unmarshal response body: %s\n", err)
		return nil, err
	}

	return labelResponse.Results, nil
}
