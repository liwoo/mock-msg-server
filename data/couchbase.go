package data

import (
	"encoding/json"
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/couchbase/gocb/v2"
	"log"
	"mock-server/cliff"
	"mock-server/shared"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	CouchbaseURL      string
	CouchbaseReadsDB  string
	CouchbaseWritesDB string
	CouchbaseUser     string
	CouchbasePass     string
	SampleDistrictId  int
	Cluster           *gocb.Cluster
	ReadsBucket       *gocb.Bucket
	WritesBucket      *gocb.Bucket
}

type ClientGroup struct {
	Id     int    `json:"id"`
	Name   string `json:"name" faker:"name"`
	Leader string `json:"leader" faker:"name"`
}

type Location struct {
	Longitude float32 `json:"longitude" faker:"long"`
	Latitude  float32 `json:"latitude" faker:"lat"`
}

type Contact struct {
	PrimaryPhoneNumber   string      `json:"primaryPhoneNumber" faker:"e_164_phone_number"`
	SecondaryPhoneNumber interface{} `json:"secondaryPhoneNumber" faker:"e_164_phone_number"`
	Email                interface{} `json:"email" faker:"email"`
	AddressLine1         interface{} `json:"addressLine1" faker:"sentence"`
	AddressLine2         interface{} `json:"addressLine2" faker:"sentence"`
}

type DocumentState struct {
	Time   time.Time `json:"time"`
	Status string    `json:"status"`
}

type ApiRequest struct {
	ClientEmail        string          `json:"clientEmail"`
	CreatedAt          time.Time       `json:"createdAt"`
	Endpoint           string          `json:"endpoint"`
	Headers            string          `json:"headers"`
	Id                 string          `json:"id"`
	OrganizationUnits  string          `json:"organizationUnits"`
	RequestData        string          `json:"requestData"`
	Type               string          `json:"type"`
	Verb               string          `json:"verb"`
	ResponseStatusCode int             `json:"responseStatusCode"`
	ResponseData       string          `json:"responseData"`
	ClientMetadata     string          `json:"clientMetadata"`
	Checksum           string          `json:"checksum"`
	DocumentStates     []DocumentState `json:"documentStates"`
}

type Client struct {
	Id               string      `json:"_id"`
	AccountNo        string      `json:"accountNo" faker:"cc_number"`
	Active           bool        `json:"active"`
	ActivationDate   []string    `json:"activationDate" faker:"timestamp, slice_len=1"`
	Firstname        string      `json:"firstname" faker:"first_name"`
	Lastname         string      `json:"lastname" faker:"last_name"`
	DisplayName      string      `json:"displayName" faker:"name"`
	OfficeId         int         `json:"officeId"`
	Dob              string      `json:"dob" faker:"date"`
	Gender           string      `json:"gender" faker:"oneof: M, F"`
	NationalIdNumber string      `json:"nationalIdNumber" faker:"cc_number"`
	Location         Location    `json:"location"`
	Contacts         Contact     `json:"contacts"`
	Group            ClientGroup `json:"group"`
	Channels         []string    `json:"channels"`
	SyncTs           string      `json:"syncTs" faker:"date"`
	Type             string      `json:"type" faker:"oneof: clients"`
}

type GroupConfigurations struct {
	MinClientsInGroup int `json:"minClientsInGroup"`
	MaxClientsInGroup int `json:"maxClientsInGroup"`
}

type Group struct {
	Id             string              `json:"id"`
	AccountNo      string              `json:"accountNo" faker:"cc_number"`
	Name           string              `json:"name" faker:"name"`
	Active         bool                `json:"active"`
	ActivationDate []string            `json:"activationDate" faker:"timestamp, slice_len=1"`
	OfficeId       int                 `json:"officeId"`
	OfficeName     string              `json:"officeName" faker:"name"`
	Channels       []string            `json:"channels"`
	Configurations GroupConfigurations `json:"configurations"`
	SyncTs         string              `json:"syncTs" faker:"date"`
	Type           string              `json:"type" faker:"oneof: groups"`
}

type ClientUpdateDto struct {
	AccountNumber string    `json:"accountNumber"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func NewService(
	couchbaseURL string,
	couchbaseReadsDB string,
	couchbaseWritesDB string,
	couchbaseUser string,
	couchbasePass string,
) *Service {
	return &Service{
		CouchbaseURL:      couchbaseURL,
		CouchbaseReadsDB:  couchbaseReadsDB,
		CouchbaseWritesDB: couchbaseWritesDB,
		CouchbaseUser:     couchbaseUser,
		CouchbasePass:     couchbasePass,
	}
}

func (s *Service) ensureConnection() error {

	if s.Cluster != nil && s.ReadsBucket != nil {
		return nil
	}

	//gocb.SetLogger(gocb.VerboseStdioLogger())

	cluster, err := gocb.Connect("couchbase://"+s.CouchbaseURL, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: s.CouchbaseUser,
			Password: s.CouchbasePass,
		},
	})

	if err != nil {
		return err
	}

	readsBucket := cluster.Bucket(s.CouchbaseReadsDB)
	writesBucket := cluster.Bucket(s.CouchbaseWritesDB)

	err = readsBucket.WaitUntilReady(5*time.Second, nil)

	if err != nil {
		return err
	}

	log.Println("Successfully connected to Couchbase")

	s.Cluster = cluster
	s.ReadsBucket = readsBucket
	s.WritesBucket = writesBucket

	return nil
}

func (s *Service) ProcessApiRequest(id string, cliffService *cliff.Service) error {
	err := s.ensureConnection()

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Processing API request for", id)
	wCol := s.WritesBucket.DefaultCollection()
	//wCol := s.WritesBucket.DefaultCollection()

	results, err := wCol.Get(id, nil)

	if err != nil {
		log.Println(err)
		return err
	}

	var apiRequest ApiRequest
	err = results.Content(&apiRequest)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Found document with id", id)

	apiRequest.DocumentStates = []DocumentState{
		{
			Time:   time.Now(),
			Status: "PROCESSING",
		},
	}

	_, err = wCol.Upsert(id, apiRequest, nil)

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Updated Document State for ", id)

	go func() {
		log.Println("Started Processing Document", id)

		var parsedClientRequestBody shared.ParsedClientRequestBody
		err1 := json.Unmarshal([]byte(apiRequest.RequestData), &parsedClientRequestBody)

		apiRequest.ResponseStatusCode = 201

		newStates := append(apiRequest.DocumentStates, DocumentState{
			Time:   time.Now(),
			Status: "PROCESSED",
		})
		apiRequest.DocumentStates = newStates

		if err1 != nil {
			log.Println(err)
			apiRequest.ResponseStatusCode = 400
		}

		resp, err2 := cliffService.CreateClient(parsedClientRequestBody)

		if err2 != nil {
			log.Println(err)
			apiRequest.ResponseStatusCode = 500
		}

		if err1 == nil && err2 == nil {
			clientUpdateResonse := ClientUpdateDto{
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				AccountNumber: resp.AccountNo,
			}
			response, _ := json.Marshal(clientUpdateResonse)
			apiRequest.ResponseData = string(response)
		}
		
		_, err = wCol.Upsert(id, apiRequest, nil)

		if err != nil {
			log.Println(err)
		}

		log.Println("Finished Processing Document", id)
	}()

	return nil
}

func (s *Service) SaveInitialClients(cliffClients []shared.ClientDTO) {
	err := s.ensureConnection()

	if err != nil {
		log.Println(err)
		return
	}

	col := s.ReadsBucket.DefaultCollection()
	for _, client := range cliffClients {
		cbClient := convertCliffClientToClient(client)

		_, err = col.Upsert(cbClient.Id, cbClient, nil)
		log.Println("Saved client", cbClient.Id)

		if err != nil {
			log.Println("Couldn't save client", err)
			continue
		}
	}
	return
}

func (s Service) SaveInitialGroups() {

}

func (s Service) UpdateClientFromWebhook(cliffClient shared.ClientDTO) error {
	err := s.ensureConnection()

	if err != nil {
		return err
	}

	col := s.ReadsBucket.DefaultCollection()
	cbClient := convertCliffClientToClient(cliffClient)

	_, err = col.Upsert(cbClient.Id, cbClient, nil)
	log.Println("Saved client", cbClient.Id)

	if err != nil {
		return err
	}

	log.Println("Updated client", cbClient.Id)

	return nil
}

func convertCliffClientToClient(client shared.ClientDTO) Client {
	var strActivationDate []string
	for _, date := range client.ActivationDate {
		strActivationDate = append(strActivationDate, strconv.Itoa(date))
	}

	//TODO: Fix this
	//dob := time.Date(client.DateOfBirth[0], time.Month(client.DateOfBirth[1]), client.DateOfBirth[2], 0, 0, 0, 0, time.UTC)
	//to format 1981-01-01T00:00:00.000Z
	dobStr := "1991-04-01" //dob.Format(time.RFC3339)

	contacts := Contact{
		PrimaryPhoneNumber: client.MobileNo,
	}

	cbClient := Client{
		Id:               "clients_" + client.AccountNo,
		AccountNo:        client.AccountNo,
		Active:           client.Active,
		ActivationDate:   strActivationDate,
		Firstname:        client.Firstname,
		Lastname:         client.Lastname,
		DisplayName:      client.DisplayName,
		OfficeId:         client.OfficeId,
		Dob:              dobStr,
		Gender:           client.Gender.Name,
		NationalIdNumber: client.ExternalId,
		Contacts:         contacts,
		Channels:         []string{"clients_" + strconv.Itoa(client.OfficeId)},
		SyncTs:           time.Now().Format("2006-01-02 15:04:05"),
		Type:             "clients",
	}
	return cbClient
}

func (s *Service) PublishDocs(channels []string) (uint, error) {
	err := s.ensureConnection()

	if err != nil {
		log.Println(err)
		return 0, err
	}

	//const
	const numDocs = 20
	successfulDocs := uint(0)

	//TODO: make this concurrent
	for i := 0; i < numDocs; i++ {
		col := s.ReadsBucket.DefaultCollection()
		client := Client{}
		group := Group{}

		err = faker.FakeData(&client)
		groupError := faker.FakeData(&group)

		if err != nil {
			message := fmt.Sprintf("%s on document # %s", err, string(rune(i)))
			log.Println(message)
			continue
		}

		if groupError != nil {
			message := fmt.Sprintf("%s on document # %s", err, string(rune(i)))
			log.Println(message)
			continue
		}

		//TODO: get channel with underscore in it...
		//find channel with "clients" in it
		clientChannel, groupChannel := "", ""
		for _, channel := range channels {
			if strings.Contains(channel, "clients") {
				clientChannel = channel
			}
			if strings.Contains(channel, "groups") {
				groupChannel = channel
			}
		}

		client.Channels = []string{clientChannel}
		group.Channels = []string{groupChannel}
		now := time.Now().Format("2006-01-02 15:04:05")
		client.SyncTs = now
		group.SyncTs = now
		client.Id = "clients_" + client.AccountNo
		group.Id = "groups_" + group.AccountNo
		//unix timestamp
		client.ActivationDate = []string{fmt.Sprintf("%d", time.Now().Unix())}

		_, err = col.Upsert(client.Id, client, nil)
		_, groupError = col.Upsert(group.Id, group, nil)

		if err != nil {
			message := fmt.Sprintf("%s on document # %s", err, string(rune(i)))
			log.Println(message)
			continue
		} else {
			log.Println("Successfully published Client document ID", client.Id)
			successfulDocs += 1
		}

		if groupError != nil {
			message := fmt.Sprintf("%s on document # %s", err, string(rune(i)))
			log.Println(message)
			continue
		} else {
			log.Println("Successfully published Group document ID", group.Id)
			successfulDocs += 1
		}
	}
	return successfulDocs, nil
}
