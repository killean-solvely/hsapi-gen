package go_example_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

/****** HTTP Client ******/

var baseURL = "https://api.hubspot.com"

type customTransport struct {
	apikey string
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+c.apikey)
	return http.DefaultTransport.RoundTrip(clonedReq)
}

func NewHubspotHTTPClient(token string) *http.Client {
	client := &http.Client{
		Transport: &customTransport{
			apikey: token,
		},
	}

	return client
}

func apiFunction(
	client *http.Client,
	method string,
	path string,
	body io.Reader,
) (*http.Response, error) {
	url := baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func apiFunctionWithResponse[T any](
	client *http.Client,
	method string,
	path string,
	body io.Reader,
) (T, error) {
	var obj T

	resp, err := apiFunction(client, method, path, body)
	if err != nil {
		return obj, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return obj, err
	}

	err = json.Unmarshal(respBody, &obj)
	if err != nil {
		return obj, err
	}

	return obj, nil
}

/****** Objects and properties ******/

type DealCommercialProjectEnum string

const (
	DealCommercialProjectYes DealCommercialProjectEnum = "true"
	DealCommercialProjectNo  DealCommercialProjectEnum = "false"
)

type DealProperty string

const (
	DealHSObjectID        DealProperty = "hs_object_id"
	DealHubspotOwnerID    DealProperty = "hubspot_owner_id"
	DealCommercialProject DealProperty = "commercial_project"
)

type Deal struct {
	HSObjectID        string                    `json:"hs_object_id,omitempty"`
	HubspotOwnerID    string                    `json:"hubspot_owner_id,omitempty"`
	CommercialProject DealCommercialProjectEnum `json:"commercial_project,omitempty"`
}

type DealPartial struct {
	HSObjectID        *string                    `json:"hs_object_id,omitempty"`
	HubspotOwnerID    *string                    `json:"hubspot_owner_id,omitempty"`
	CommercialProject *DealCommercialProjectEnum `json:"commercial_project,omitempty"`
}

type ContactProperty string

const (
	ContactHSObjectID ContactProperty = "hs_object_id"
	ContactFirstName  ContactProperty = "firstname"
	ContactLastName   ContactProperty = "lastname"
)

type Contact struct {
	HSObjectID string `json:"hs_object_id,omitempty"`
	FirstName  string `json:"firstname,omitempty"`
	LastName   string `json:"lastname,omitempty"`
}

type ContactPartial struct {
	HSObjectID *string `json:"hs_object_id,omitempty"`
	FirstName  *string `json:"firstname,omitempty"`
	LastName   *string `json:"lastname,omitempty"`
}

type ObjectInternalName string

const (
	DealInternalName    ObjectInternalName = "deal"
	ContactInternalName ObjectInternalName = "contact"
)

var ObjectTypeToID = map[ObjectInternalName]string{
	DealInternalName:    "0",
	ContactInternalName: "1",
}

/****** API function types and builders ******/

type HistoryEntry struct {
	Value      string    `json:"value,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	SourceType string    `json:"sourceType,omitempty"`
	SourceID   string    `json:"sourceId,omitempty"`
}

type ObjectWithHistory struct {
	Value   string
	History HistoryEntry
}

type UpdateBatchInput struct {
	ID    string
	Props map[string]string
}

type getFuncType[TObject any, TObjectProperties any] func(id string, properties []TObjectProperties) (*TObject, error)

type getWithHistoryType[TObject any, TObjectProperties ~string] func(id string, properties []TObjectProperties) (map[TObjectProperties]ObjectWithHistory, error)

type getBatchFuncType[TObject any, TObjectProperties ~string] func(ids []string, properties []TObjectProperties) ([]TObject, error)

type createFuncType[TObject any, TPartialObject any] func(obj TPartialObject) (*TPartialObject, error)

type createBatchFuncType[TObject any, TPartialObject any] func(objs []TPartialObject) ([]TObject, error)

type updateFuncType[TObject any, TObjectProperties ~string] func(id string, props map[TObjectProperties]string) error

type updateBatchFuncType[TObject any, TObjectProperties ~string] func(inputs []UpdateBatchInput) error

type getAssociationsFuncType[TObject any] func(fromObjID string, toObjType ObjectInternalName) ([]string, error)

type associateFuncType func(fromObjID string, toObjID string) error

/****** Actual API schemas using the API types ******/

type GetFunc struct {
	Get struct {
		Contact getFuncType[Contact, ContactProperty]
		Deal    getFuncType[Deal, DealProperty]
	}
}

type GetWithHistoryFunc struct {
	GetWithHistory struct {
		Contact getWithHistoryType[Contact, ContactProperty]
		Deal    getWithHistoryType[Deal, DealProperty]
	}
}

type GetBatchFunc struct {
	GetBatch struct {
		Contact getBatchFuncType[Contact, ContactProperty]
		Deal    getBatchFuncType[Deal, DealProperty]
	}
}

type CreateFunc struct {
	Create struct {
		Contact createFuncType[Contact, ContactPartial]
		Deal    createFuncType[Deal, DealPartial]
	}
}

type CreateBatchFunc struct {
	CreateBatch struct {
		Contact createBatchFuncType[Contact, ContactPartial]
		Deal    createBatchFuncType[Deal, DealPartial]
	}
}

type UpdateFunc struct {
	Update struct {
		Contact updateFuncType[Contact, ContactProperty]
		Deal    updateFuncType[Deal, DealProperty]
	}
}

type UpdateBatchFunc struct {
	UpdateBatch struct {
		Contact updateBatchFuncType[Contact, ContactProperty]
		Deal    updateBatchFuncType[Deal, DealProperty]
	}
}

type GetAssociationsFunc struct {
	GetAssociations struct {
		Contact getAssociationsFuncType[Contact]
		Deal    getAssociationsFuncType[Deal]
	}
}

type AssociateFunc struct {
	Associate struct {
		Contact struct {
			Deal struct {
				ContactToDeal associateFuncType
				Special       associateFuncType
			}
		}
		Deal struct {
			Contact struct {
				DealToContact associateFuncType
			}
		}
	}
}

type APIFunc struct {
	GetFunc
	GetWithHistoryFunc
	GetBatchFunc
	CreateFunc
	CreateBatchFunc
	UpdateFunc
	UpdateBatchFunc
	GetAssociationsFunc
	AssociateFunc
}

/****** API Builder, this will handle the hubspot API calls ******/

type APIBuilder struct {
	client *http.Client
}

func NewAPIBuilder(token string) *APIBuilder {
	return &APIBuilder{
		client: NewHubspotHTTPClient(token),
	}
}

func getFuncBuilder[TObject any, TObjectProperties ~string](
	client *http.Client,
	objectType ObjectInternalName,
) getFuncType[TObject, TObjectProperties] {
	return func(id string, properties []TObjectProperties) (*TObject, error) {
		objInternalID := ObjectTypeToID[objectType]

		type GetResponse struct {
			ID                    string                             `json:"id,omitempty"`
			Properties            TObject                            `json:"properties,omitempty"`
			PropertiesWithHistory map[TObjectProperties]HistoryEntry `json:"propertiesWithHistory,omitempty"`
			CreatedAt             time.Time                          `json:"createdAt,omitempty"`
			UpdatedAt             time.Time                          `json:"updatedAt,omitempty"`
			Archived              bool                               `json:"archived,omitempty"`
		}

		propertiesQuery := ""
		if len(properties) > 0 {
			propertiesQuery = "?properties=" + string(properties[0])
			for _, prop := range properties[1:] {
				propertiesQuery += "&properties=" + string(prop)
			}
		}

		resp, err := apiFunctionWithResponse[GetResponse](
			client,
			"GET",
			"/crm/v3/objects/"+objInternalID+"/"+id+propertiesQuery,
			nil,
		)
		if err != nil {
			return nil, err
		}

		return &resp.Properties, nil
	}
}

func getWithHistoryFuncBuilder[TObject any, TObjectProperties ~string](
	client *http.Client,
	objectType ObjectInternalName,
) getWithHistoryType[TObject, TObjectProperties] {
	return func(id string, properties []TObjectProperties) (map[TObjectProperties]ObjectWithHistory, error) {
		objInternalID := ObjectTypeToID[objectType]

		type GetWithHistoryResponse struct {
			ID                    string                             `json:"id,omitempty"`
			Properties            map[TObjectProperties]string       `json:"properties,omitempty"`
			PropertiesWithHistory map[TObjectProperties]HistoryEntry `json:"propertiesWithHistory,omitempty"`
			CreatedAt             time.Time                          `json:"createdAt,omitempty"`
			UpdatedAt             time.Time                          `json:"updatedAt,omitempty"`
			Archived              bool                               `json:"archived,omitempty"`
		}

		resp, err := apiFunctionWithResponse[GetWithHistoryResponse](
			client,
			"GET",
			"/crm/v3/objects/"+objInternalID+"/"+id,
			nil,
		)
		if err != nil {
			return nil, err
		}

		objWithHistory := make(map[TObjectProperties]ObjectWithHistory)
		for _, prop := range properties {
			objWithHistory[prop] = ObjectWithHistory{
				Value:   resp.Properties[prop],
				History: resp.PropertiesWithHistory[prop],
			}
		}

		return objWithHistory, nil
	}
}

func getBatchFuncBuilder[TObject any, TObjectProperties ~string](
	client *http.Client,
	objectType ObjectInternalName,
) getBatchFuncType[TObject, TObjectProperties] {
	return func(ids []string, properties []TObjectProperties) ([]TObject, error) {
		type GetBatchInput struct {
			Inputs []struct {
				ID string `json:"id,omitempty"`
			} `json:"inputs,omitempty"`
			Properties []TObjectProperties `json:"properties,omitempty"`
		}

		type GetBatchResponse struct {
			Results []struct {
				ID                    string                             `json:"id,omitempty"`
				Properties            TObject                            `json:"properties,omitempty"`
				PropertiesWithHistory map[TObjectProperties]HistoryEntry `json:"propertiesWithHistory,omitempty"`
				CreatedAt             time.Time                          `json:"createdAt,omitempty"`
				UpdatedAt             time.Time                          `json:"updatedAt,omitempty"`
				Archived              bool                               `json:"archived,omitempty"`
			} `json:"results,omitempty"`
		}

		objInternalID := ObjectTypeToID[objectType]

		inputs := make([]struct {
			ID string `json:"id,omitempty"`
		}, len(ids))

		for i, id := range ids {
			inputs[i].ID = id
		}

		input := GetBatchInput{
			Inputs:     inputs,
			Properties: properties,
		}

		inputJSON, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}

		resp, err := apiFunctionWithResponse[GetBatchResponse](
			client,
			"POST",
			"/crm/v3/objects/"+objInternalID+"/batch/read",
			bytes.NewBuffer(inputJSON),
		)
		if err != nil {
			return nil, err
		}

		objs := make([]TObject, len(resp.Results))
		for i, result := range resp.Results {
			objs[i] = result.Properties
		}

		return objs, nil
	}
}

func createFuncBuilder[TObject any, TPartialObject any](
	client *http.Client,
	objectType ObjectInternalName,
) createFuncType[TObject, TPartialObject] {
	return func(obj TPartialObject) (*TPartialObject, error) {
		type CreateInput struct {
			Associations []struct{}     `json:"associations,omitempty"`
			Properties   TPartialObject `json:"properties,omitempty"`
		}
		type CreateResponse struct {
			ID         string         `json:"id,omitempty"`
			Properties TPartialObject `json:"properties,omitempty"`
			CreatedAt  time.Time      `json:"createdAt,omitempty"`
			UpdatedAt  time.Time      `json:"updatedAt,omitempty"`
		}

		objInternalID := ObjectTypeToID[objectType]

		input := CreateInput{
			Properties:   obj,
			Associations: []struct{}{},
		}

		inputJSON, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}

		resp, err := apiFunctionWithResponse[CreateResponse](
			client,
			"POST",
			"/crm/v3/objects/"+objInternalID,
			bytes.NewBuffer(inputJSON),
		)
		if err != nil {
			return nil, err
		}

		return &resp.Properties, nil
	}
}

func createBatchFuncBuilder[TObject any, TPartialObject any](
	client *http.Client,
	objectType ObjectInternalName,
) createBatchFuncType[TObject, TPartialObject] {
	return func(objs []TPartialObject) ([]TObject, error) {
		type CreateBatchInput struct {
			Inputs []struct {
				Associations []struct{}     `json:"associations,omitempty"`
				Properties   TPartialObject `json:"properties,omitempty"`
			} `json:"inputs,omitempty"`
		}
		type CreateBatchResponse struct {
			Results []struct {
				ID         string         `json:"id,omitempty"`
				Properties TPartialObject `json:"properties,omitempty"`
				CreatedAt  time.Time      `json:"createdAt,omitempty"`
				UpdatedAt  time.Time      `json:"updatedAt,omitempty"`
			} `json:"results,omitempty"`
		}

		objInternalID := ObjectTypeToID[objectType]

		inputs := make([]struct {
			Associations []struct{}     `json:"associations,omitempty"`
			Properties   TPartialObject `json:"properties,omitempty"`
		}, len(objs))

		for i, obj := range objs {
			inputs[i].Properties = obj
			inputs[i].Associations = []struct{}{}
		}

		input := CreateBatchInput{
			Inputs: inputs,
		}

		inputJSON, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}

		resp, err := apiFunctionWithResponse[CreateBatchResponse](
			client,
			"POST",
			"/crm/v3/objects/"+objInternalID+"/batch/create",
			bytes.NewBuffer(inputJSON),
		)
		if err != nil {
			return nil, err
		}

		createdObjs := make([]TObject, len(resp.Results))
		for i, result := range resp.Results {
			objs[i] = result.Properties
		}

		return createdObjs, nil
	}
}

func updateFuncBuilder[TObject any, TObjectProperties ~string](
	client *http.Client,
) updateFuncType[TObject, TObjectProperties] {
	return func(id string, props map[TObjectProperties]string) error {
		type UpdateInput struct {
			Properties map[TObjectProperties]string `json:"properties,omitempty"`
		}
		panic("not implemented")
	}
}

func updateBatchFuncBuilder[TObject any, TObjectProperties ~string](
	client *http.Client,
) updateBatchFuncType[TObject, TObjectProperties] {
	return func(inputs []UpdateBatchInput) error {
		panic("not implemented")
	}
}

func getAssociationsFuncBuilder[TObject any](client *http.Client) getAssociationsFuncType[TObject] {
	return func(fromObjID string, toObjType ObjectInternalName) ([]string, error) {
		panic("not implemented")
	}
}

func associateFuncBuilder(
	client *http.Client,
	fromObj ObjectInternalName,
	toObj ObjectInternalName,
	assocTypeID string,
) associateFuncType {
	return func(fromObjID string, toObjID string) error {
		panic("not implemented")
	}
}

/****** API function builders ******/

func (api APIBuilder) newGetFunc() GetFunc {
	var get GetFunc
	get.Get.Contact = getFuncBuilder[Contact, ContactProperty](api.client, ContactInternalName)
	get.Get.Deal = getFuncBuilder[Deal, DealProperty](api.client, DealInternalName)
	return get
}

func (api APIBuilder) newGetWithHistoryFunc() GetWithHistoryFunc {
	var get GetWithHistoryFunc
	get.GetWithHistory.Contact = getWithHistoryFuncBuilder[Contact, ContactProperty](
		api.client,
		ContactInternalName,
	)
	get.GetWithHistory.Deal = getWithHistoryFuncBuilder[Deal, DealProperty](
		api.client,
		DealInternalName,
	)
	return get
}

func (api APIBuilder) newGetBatchFunc() GetBatchFunc {
	var get GetBatchFunc
	get.GetBatch.Contact = getBatchFuncBuilder[Contact, ContactProperty](
		api.client,
		ContactInternalName,
	)
	get.GetBatch.Deal = getBatchFuncBuilder[Deal, DealProperty](api.client, DealInternalName)
	return get
}

func (api APIBuilder) newCreateFunc() CreateFunc {
	var create CreateFunc
	create.Create.Contact = createFuncBuilder[Contact, ContactPartial](
		api.client,
		ContactInternalName,
	)
	create.Create.Deal = createFuncBuilder[Deal, DealPartial](api.client, DealInternalName)
	return create
}

func (api APIBuilder) newCreateBatchFunc() CreateBatchFunc {
	var create CreateBatchFunc
	create.CreateBatch.Contact = createBatchFuncBuilder[Contact, ContactPartial](
		api.client,
		ContactInternalName,
	)
	create.CreateBatch.Deal = createBatchFuncBuilder[Deal, DealPartial](
		api.client,
		DealInternalName,
	)
	return create
}

func (api APIBuilder) newUpdateFunc() UpdateFunc {
	var update UpdateFunc
	update.Update.Contact = updateFuncBuilder[Contact, ContactProperty](api.client)
	update.Update.Deal = updateFuncBuilder[Deal, DealProperty](api.client)
	return update
}

func (api APIBuilder) newUpdateBatchFunc() UpdateBatchFunc {
	var update UpdateBatchFunc
	update.UpdateBatch.Contact = updateBatchFuncBuilder[Contact, ContactProperty](api.client)
	update.UpdateBatch.Deal = updateBatchFuncBuilder[Deal, DealProperty](api.client)
	return update
}

func (api APIBuilder) newGetAssociationsFunc() GetAssociationsFunc {
	var get GetAssociationsFunc
	get.GetAssociations.Contact = getAssociationsFuncBuilder[Contact](api.client)
	get.GetAssociations.Deal = getAssociationsFuncBuilder[Deal](api.client)
	return get
}

func (api APIBuilder) newAssociateFunc() AssociateFunc {
	var associate AssociateFunc
	associate.Associate.Contact.Deal.ContactToDeal = associateFuncBuilder(
		api.client,
		ContactInternalName,
		DealInternalName,
		"0",
	)
	associate.Associate.Contact.Deal.Special = associateFuncBuilder(
		api.client,
		ContactInternalName,
		DealInternalName,
		"1",
	)
	associate.Associate.Deal.Contact.DealToContact = associateFuncBuilder(
		api.client,
		DealInternalName,
		ContactInternalName,
		"3",
	)
	return associate
}

func newAPIFunc(token string) APIFunc {
	var api APIFunc

	apiBuilder := NewAPIBuilder(token)
	api.GetFunc = apiBuilder.newGetFunc()
	api.GetWithHistoryFunc = apiBuilder.newGetWithHistoryFunc()
	api.GetBatchFunc = apiBuilder.newGetBatchFunc()
	api.CreateFunc = apiBuilder.newCreateFunc()
	api.CreateBatchFunc = apiBuilder.newCreateBatchFunc()
	api.UpdateFunc = apiBuilder.newUpdateFunc()
	api.UpdateBatchFunc = apiBuilder.newUpdateBatchFunc()
	api.GetAssociationsFunc = apiBuilder.newGetAssociationsFunc()
	api.AssociateFunc = apiBuilder.newAssociateFunc()

	return api
}

/****** Client ******/

type HubspotClient struct {
	API APIFunc
}

func NewHubspotClient(token string) HubspotClient {
	return HubspotClient{
		API: newAPIFunc(token),
	}
}

func tryClient() {
	c := NewHubspotClient("token")
	deal, err := c.API.Get.Deal(
		"dealid",
		[]DealProperty{DealHSObjectID, DealHubspotOwnerID},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Deal: %#v\n", deal)
}
