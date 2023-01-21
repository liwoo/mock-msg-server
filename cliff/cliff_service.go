package cliff

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"mock-server/shared"
	"net/http"
	"strconv"
	"time"
)

type Service struct {
	BaseURL string
	Token   string
	//For Demo Purposes
	DefaultOfficeId      string
	GetClientsEndpoint   string
	GetGroupsEndpoint    string
	CreateClientEndpoint string
	UpdateClientEndpoint string
}

type WebhookRequestOffice struct {
	OfficeId float64 `json:"officeId"`
}

type WebhookRequest struct {
	Locale                  string                 `json:"locale"`
	DateFormat              string                 `json:"dateFormat"`
	Name                    string                 `json:"name"`
	FromDate                string                 `json:"fromDate"`
	ToDate                  string                 `json:"toDate"`
	ReschedulingType        float64                `json:"reschedulingType"`
	RepaymentsRescheduledTo string                 `json:"repaymentsRescheduledTo"`
	Description             string                 `json:"description"`
	Offices                 []WebhookRequestOffice `json:"offices"`
}

type WebhookResponse struct {
	ResourceId int `json:"resourceId"`
}

type WebhookPayload struct {
	CreatedByName     string          `json:"createdByName"`
	Request           WebhookRequest  `json:"request"`
	CreatedBy         int             `json:"createdBy"`
	EntityName        string          `json:"entityName"`
	Response          WebhookResponse `json:"response"`
	CreatedByFullName string          `json:"createdByFullName"`
	ActionName        string          `json:"actionName"`
	Timestamp         time.Time       `json:"timestamp"`
}

func NewCliffService(
	baseUrl string,
	token string,
	defaultOfficeId string,
	getClientsEndpoint string,
	getGroupsEndpoint string,
	createClientEndpoint string,
	updateClientEndpoint string,

) *Service {
	return &Service{
		BaseURL:              baseUrl,
		Token:                token,
		DefaultOfficeId:      defaultOfficeId,
		GetClientsEndpoint:   getClientsEndpoint,
		GetGroupsEndpoint:    getGroupsEndpoint,
		CreateClientEndpoint: createClientEndpoint,
		UpdateClientEndpoint: updateClientEndpoint,
	}
}

func getCliffRequest(url string, method string, token string) (*http.Request, error) {
	httpRequest, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	httpRequest.Header.Add("Authorization", "Bearer "+token)
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("fineract-platform-tenantid", "default")
	return httpRequest, nil
}

func (s Service) GetClientById(clientId string) (shared.ClientDTO, error) {
	request, err := getCliffRequest(s.BaseURL+s.GetClientsEndpoint+"/"+clientId, "GET", s.Token)
	if err != nil {
		log.Println(err)
		return shared.ClientDTO{}, err
	}

	client := http.Client{}

	resp, err := client.Do(request)

	if err != nil {
		log.Println(err)
		return shared.ClientDTO{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return shared.ClientDTO{}, err
	}

	var clientResponse shared.ClientDTO
	err = json.Unmarshal(body, &clientResponse)

	if err != nil {
		log.Println(err)
		return shared.ClientDTO{}, err
	}

	return clientResponse, nil
}

func (s Service) UpsertClient(body shared.ParsedClientRequestBody, method string) (shared.CreateClientResponse, error, int) {
	cliffClientRequestCreate := convertCbClientToCliffClient(body, s.DefaultOfficeId)
	//cliffClientRequestUpdate := convertCbClientToCliffUpdateClient(body, s.DefaultOfficeId)

	cliffClientRequestCreateBody, err := json.Marshal(cliffClientRequestCreate)
	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 400
	}
	//cliffClientRequestUpdateBody, err := json.Marshal(cliffClientRequestUpdate)
	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 400
	}

	//clientId := "45"
	log.Println("Cliff Client Request Body: ", string(cliffClientRequestCreateBody))
	url := s.BaseURL + s.CreateClientEndpoint
	//if method == "PUT" {
	//	url = s.BaseURL + s.UpdateClientEndpoint + "/" + clientId
	//}
	request, err := getCliffRequest(url, method, s.Token)

	bodyBuffer := bytes.NewBuffer(cliffClientRequestCreateBody)
	//if method == "PUT" {
	//	bodyBuffer = bytes.NewBuffer(cliffClientRequestUpdateBody)
	//}
	request.Body = ioutil.NopCloser(bodyBuffer)

	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 400
	}

	client := http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 400
	}

	//read response status code\
	statusCode := resp.StatusCode

	//handle bad status code
	if statusCode != 200 && statusCode != 201 {
		log.Println("Bad Status Code: ", statusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response body: ", err)
		}
		log.Println("Response Body: ", string(body))
		return shared.CreateClientResponse{}, errors.New(string(body)), statusCode
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 500
	}

	var response shared.CreateClientResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Println(err)
		return shared.CreateClientResponse{}, err, 500
	}

	return response, nil, 200
}

func convertCbClientToCliffUpdateClient(body shared.ParsedClientRequestBody, officeId string) shared.ClientUpdateBody {
	longitude := "0.0"
	latitude := "0.0"

	address := shared.ClientAddress{
		Street:       "N/A",
		AddressLine1: "N/A",
		AddressLine2: "N/A",
		CloseTown:    "N/A",
		VillageName:  "N/A",
		CellName:     "N/A",
		PostalCode:   "N/A",
		City:         "N/A",
		Latitude:     longitude,
		Longitude:    latitude,
		Locale:       body.ClientBio.Locale,
	}
	//todays date in format 27 January 2022
	return shared.ClientUpdateBody{
		ClientAddress:  address,
		LegalFormId:    1,
		Firstname:      body.ClientBio.Firstname,
		Lastname:       body.ClientBio.Lastname,
		MobileNo:       body.ClientBio.PrimaryPhoneNumber,
		Locale:         body.ClientBio.Locale,
		Active:         true,
		DateFormat:     "dd MMMM yyyy",
		ActivationDate: "05 January 2022",
	}
}

func convertCbClientToCliffClient(body shared.ParsedClientRequestBody, officeId string) shared.CreateClientDTO {
	officeIdInt, err := strconv.Atoi(officeId)
	if err != nil {
		log.Println(err)
		officeIdInt = 240
	}
	identifier := shared.ClientIdentifier{
		DocumentTypeId: body.ClientId.DocumentTypeId,
		DocumentKey:    body.ClientId.DocumentKey,
		Status:         "ACTIVE",
	}
	longitude := "0.0"
	latitude := "0.0"

	address := shared.ClientAddress{
		Street:       "N/A",
		AddressLine1: "N/A",
		AddressLine2: "N/A",
		CloseTown:    "N/A",
		VillageName:  "N/A",
		CellName:     "N/A",
		PostalCode:   "N/A",
		City:         "N/A",
		Latitude:     longitude,
		Longitude:    latitude,
		Locale:       body.ClientBio.Locale,
	}
	//todays date in format 27 January 2022
	return shared.CreateClientDTO{
		ClientAddress:  address,
		OfficeId:       officeIdInt,
		FamilyMembers:  []interface{}{},
		LegalFormId:    1,
		Firstname:      body.ClientBio.Firstname,
		Lastname:       body.ClientBio.Lastname,
		MobileNo:       body.ClientBio.PrimaryPhoneNumber,
		Locale:         body.ClientBio.Locale,
		Active:         true,
		DateFormat:     "dd MMMM yyyy",
		DateOfBirth:    "01 January 1990",
		ActivationDate: "05 January 2022",
		Identifiers:    []shared.ClientIdentifier{identifier},
	}
}

func (s *Service) GetOfficeClients(officeId string) ([]shared.ClientDTO, error) {

	client := &http.Client{}
	url := s.BaseURL + s.GetClientsEndpoint + "?officeId=" + officeId
	httpRequest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	httpRequest.Header.Add("Authorization", "Bearer "+s.Token)
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("fineract-platform-tenantid", "default")

	//perform request
	httpResponse, err := client.Do(httpRequest)

	//handle error
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer httpResponse.Body.Close()

	//unmarshal response
	var response shared.ClientsResponseDTO
	err = json.NewDecoder(httpResponse.Body).Decode(&response)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	//handle response
	clients := response.PageItems

	return clients, nil
}
