package auth

import (
	"encoding/json"
	"errors"
	//"github.com/golang-jwt/jwt"
	"github.com/bxcodec/faker/v3"
	uuid2 "github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ClaimStrings []string

type Service struct {
	JWT        string
	Secret     string
	SGWBaseURL string
	DistrictId int
	CountryId  int
}

type OafClaims struct {
	GivenName string   `json:"given_name" faker:"first_name"`
	Surname   string   `json:"family_name" faker:"last_name"`
	Email     string   `json:"email" faker:"email"`
	Role      []string `json:"roles"`
}

type SGWRequest struct {
	Name          string   `json:"name"`
	Password      string   `json:"password"`
	AdminChannels []string `json:"admin_channels"`
	AllChannels   []string `json:"all_channels"`
	Disabled      bool     `json:"disabled"`
	AdminRoles    []string `json:"admin_roles"`
	Roles         []string `json:"roles"`
}

type OU struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Parent    int    `json:"parent"`
	LevelName string `json:"level_name"`
	IsCountry bool   `json:"is_country"`
}

type SGWResponse struct {
	Name           string   `json:"name"`
	Password       string   `json:"password"`
	AdminRoles     []string `json:"admin_roles"`
	AdminChannels  []string `json:"admin_channels"`
	GeographicInfo []OU     `json:"geographic_info"`
}

type CustomClaims struct {
	Audience ClaimStrings `json:"aud,omitempty"`
	//*jwt.StandardClaims
	OafClaims
}

func (c CustomClaims) Valid() error {
	return nil
}

func (s Service) NewService(
	jwt string,
	secret string,
	sGWBaseURL string,
) *Service {
	return &Service{
		JWT:        jwt,
		Secret:     secret,
		SGWBaseURL: sGWBaseURL,
	}
}

func (s Service) generateOrgUnits() []OU {
	var orgUnits []OU
	countries := []string{"Kenya", "Rwanda"}
	//random number of org units to generate
	numOrgUnits, _ := faker.RandomInt(1, 5, 1)
	for i := 0; i < numOrgUnits[0]; i++ {
		randomId, _ := faker.RandomInt(100, 999, 1)
		//random number between 1 and 2
		randomCountry, _ := faker.RandomInt(0, len(countries)-1, 1)
		parent := 0
		name := faker.Name()
		isCountry := false
		levelName := faker.Name()
		if i == 0 {
			name = countries[randomCountry[0]]
			isCountry = true
			levelName = "Country"
		}
		if i > 0 {
			parent = orgUnits[i-1].Id
		}
		orgUnits = append(orgUnits, OU{Id: randomId[0], Name: name, IsCountry: isCountry, LevelName: levelName, Parent: parent})
	}
	return orgUnits
}

func (s Service) CreateSGWUser(claims *CustomClaims) (SGWResponse, error) {
	//replace @ with _ in email
	email := strings.Replace(claims.Email, "@", "_", -1)
	uuid := uuid2.New().String()
	var roles []string

	for _, role := range claims.Role {
		roles = append(roles, strings.Replace(role, " ", "_", -1))
	}
	//convert i to string

	var channelList []string

	generatedOus := s.generateOrgUnits()

	for _, channel := range generatedOus {
		channelList = append(channelList, strconv.Itoa(channel.Id))
	}

	joinedChannels := strings.Join(channelList, "_")
	channel := "clients_" + joinedChannels

	requestBody := SGWRequest{
		Name:          email,
		Password:      uuid,
		AdminChannels: []string{channel, strings.Replace(claims.Email, "@", "_", 1)},
		AllChannels:   []string{channel, strings.Replace(claims.Email, "@", "_", 1), "!"},
		Disabled:      false,
		AdminRoles:    []string{"replicator"},
		Roles:         append(roles, "replicator"),
	}

	//marshall the request body
	requestBodyJson, err := json.Marshal(requestBody)

	if err != nil {
		log.Println("Error marshalling request body", err)
		return SGWResponse{}, err
	}

	var requestBodyString = string(requestBodyJson)
	//make http call to SGW to create user
	createUserEndpointReads := s.SGWBaseURL + "/offline_reads/_user/"
	response, err := http.Post(createUserEndpointReads, "application/json", strings.NewReader(requestBodyString))

	if err != nil {
		log.Println(err)
		return SGWResponse{}, err
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		log.Println("Error creating user on Offline Reads", response.StatusCode)
		return SGWResponse{}, errors.New("error creating offline reads user")
	}

	createUserEndpointReads = s.SGWBaseURL + "/offline_writes/_user/"
	response, err = http.Post(createUserEndpointReads, "application/json", strings.NewReader(requestBodyString))

	if err != nil {
		log.Println(err)
		return SGWResponse{}, err
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		log.Println("Error creating user on Offline Writes", response.StatusCode)
		return SGWResponse{}, errors.New("error creating offline writes user")
	}

	responseBody := SGWResponse{
		Name:           email,
		Password:       uuid,
		AdminRoles:     claims.Role,
		AdminChannels:  []string{channel, strings.Replace(claims.Email, "@", "_", 1)},
		GeographicInfo: generatedOus,
	}

	return responseBody, nil
}

func (s Service) RetrieveClaims() (CustomClaims, error) {
	//secret := "-----BEGIN CERTIFICATE-----\n" + s.Secret + "\n-----END CERTIFICATE-----"
	//verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(secret))
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.Println("Recovered in RetrieveClaims", r)
	//	}
	//}()
	//// Parse the token
	//token, err := jwt.ParseWithClaims(s.JWT, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
	//	return verifyKey, nil
	//})
	//
	//if err != nil {
	//	return CustomClaims{}, err
	//}
	//
	//if token.Valid == false {
	//	log.Println("Token is not valid")
	//	return CustomClaims{}, errors.New("token is not valid")
	//}
	//
	//claims := token.Claims.(*CustomClaims)
	//check if the token is at least 100 chars long
	if len(s.JWT) < 100 {
		return CustomClaims{}, errors.New("token is not valid")
	}
	claims := CustomClaims{}
	roles := []string{"replicator", "oaf_dev", "oaf_fo"}

	err := faker.FakeData(&claims)

	if err != nil {
		log.Println("Error generating fake data", err)
		return CustomClaims{}, err
	}

	claims.Role = roles

	return claims, nil
}
