package shared

type ClientsResponseDTO struct {
	TotalFilteredRecords int         `json:"totalFilteredRecords"`
	PageItems            []ClientDTO `json:"pageItems"`
}

type ClientUpdateBody struct {
	ClientAddress  ClientAddress `json:"clientAddress"`
	LegalFormId    int           `json:"legalFormId"`
	Firstname      string        `json:"firstname"`
	Lastname       string        `json:"lastname"`
	MobileNo       string        `json:"mobileNo"`
	Locale         string        `json:"locale"`
	Active         bool          `json:"active"`
	DateFormat     string        `json:"dateFormat"`
	ActivationDate string        `json:"activationDate"`
}

type ClientAddress struct {
	Street       string `json:"street"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	CloseTown    string `json:"closeTown"`
	VillageName  string `json:"villageName"`
	CellName     string `json:"cellName"`
	City         string `json:"city"`
	PostalCode   string `json:"postalCode"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Locale       string `json:"locale"`
}

type ClientIdentifier struct {
	DocumentTypeId int    `json:"documentTypeId"`
	DocumentKey    string `json:"documentKey"`
	Status         string `json:"status"`
}

type CreateClientDTO struct {
	ClientAddress  ClientAddress      `json:"clientAddress"`
	FamilyMembers  []interface{}      `json:"familyMembers"`
	OfficeId       int                `json:"officeId"`
	LegalFormId    int                `json:"legalFormId"`
	Firstname      string             `json:"firstname"`
	Lastname       string             `json:"lastname"`
	MobileNo       string             `json:"mobileNo"`
	Locale         string             `json:"locale"`
	Active         bool               `json:"active"`
	DateFormat     string             `json:"dateFormat"`
	ActivationDate string             `json:"activationDate"`
	DateOfBirth    string             `json:"dateOfBirth"`
	Identifiers    []ClientIdentifier `json:"identifiers"`
}

type GroupStatus struct {
	Id    int    `json:"id"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

type GroupTimeline struct {
	SubmittedOnDate      []int  `json:"submittedOnDate"`
	SubmittedByUsername  string `json:"submittedByUsername"`
	SubmittedByFirstname string `json:"submittedByFirstname"`
	SubmittedByLastname  string `json:"submittedByLastname"`
}

type GroupConfiguration struct {
	MaxClientsInGroup int `json:"maxClientsInGroup"`
}

type GroupDTO struct {
	Id             int                `json:"id"`
	AccountNo      string             `json:"accountNo"`
	Name           string             `json:"name"`
	Status         GroupStatus        `json:"status"`
	ActivationDate []int              `json:"activationDate"`
	Active         bool               `json:"active"`
	OfficeId       int                `json:"officeId"`
	OfficeName     string             `json:"officeName"`
	Hierarchy      string             `json:"hierarchy"`
	GroupLevel     string             `json:"groupLevel"`
	Timeline       GroupTimeline      `json:"timeline"`
	Configurations GroupConfiguration `json:"configurations"`
}

type CreateClientResponse struct {
	OfficeId   int    `json:"officeId"`
	ClientId   int    `json:"clientId"`
	ResourceId int    `json:"resourceId"`
	AccountNo  string `json:"accountNo"`
	CountryId  int    `json:"countryId"`
}

type ClientDTO struct {
	Id         int    `json:"id"`
	AccountNo  string `json:"accountNo"`
	ExternalId string `json:"externalId"`
	Status     struct {
		Id    int    `json:"id"`
		Code  string `json:"code"`
		Value string `json:"value"`
	} `json:"status"`
	SubStatus struct {
		Active    bool `json:"active"`
		Mandatory bool `json:"mandatory"`
	} `json:"subStatus"`
	Active         bool   `json:"active"`
	ActivationDate []int  `json:"activationDate"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
	DisplayName    string `json:"displayName"`
	MobileNo       string `json:"mobileNo"`
	DateOfBirth    []int  `json:"dateOfBirth"`
	Gender         struct {
		Id        int    `json:"id"`
		Name      string `json:"name"`
		Active    bool   `json:"active"`
		Mandatory bool   `json:"mandatory"`
	} `json:"gender"`
	ClientType struct {
		Active    bool `json:"active"`
		Mandatory bool `json:"mandatory"`
	} `json:"clientType"`
	ClientClassification struct {
		Active    bool `json:"active"`
		Mandatory bool `json:"mandatory"`
	} `json:"clientClassification"`
	IsStaff    bool   `json:"isStaff"`
	HasLoans   bool   `json:"hasLoans"`
	OfficeId   int    `json:"officeId"`
	OfficeName string `json:"officeName"`
	Timeline   struct {
		SubmittedOnDate      []int  `json:"submittedOnDate"`
		SubmittedByUsername  string `json:"submittedByUsername"`
		SubmittedByFirstname string `json:"submittedByFirstname"`
		SubmittedByLastname  string `json:"submittedByLastname"`
		ActivatedOnDate      []int  `json:"activatedOnDate"`
		ActivatedByUsername  string `json:"activatedByUsername"`
		ActivatedByFirstname string `json:"activatedByFirstname"`
		ActivatedByLastname  string `json:"activatedByLastname"`
	} `json:"timeline"`
	LegalForm struct {
		Id    int    `json:"id"`
		Code  string `json:"code"`
		Value string `json:"value"`
	} `json:"legalForm"`
	ClientNonPersonDetails struct {
		Constitution struct {
			Active    bool `json:"active"`
			Mandatory bool `json:"mandatory"`
		} `json:"constitution"`
		MainBusinessLine struct {
			Active    bool `json:"active"`
			Mandatory bool `json:"mandatory"`
		} `json:"mainBusinessLine"`
	} `json:"clientNonPersonDetails"`
	CountryId int `json:"countryId"`
}

type ParsedClientRequestBody struct {
	ClientId      ClientBodyIdentifier `json:"clientId"`
	ClientBio     ClientBodyBio        `json:"clientBio"`
	ClientAddress ClientBodyAddress    `json:"clientAddress"`
}

type ClientBodyIdentifier struct {
	DocumentTypeId int    `json:"documentTypeId"`
	DocumentKey    string `json:"documentKey"`
	Description    string `json:"description"`
}

type ClientBodyBio struct {
	OfficeId           interface{} `json:"officeId"`
	Fullname           string      `json:"fullname"`
	Firstname          string      `json:"firstname"`
	Lastname           string      `json:"lastname"`
	GroupId            interface{} `json:"groupId"`
	DateFormat         string      `json:"dateFormat"`
	Locale             string      `json:"locale"`
	Active             bool        `json:"active"`
	ActivationDate     string      `json:"activationDate"`
	GenderId           string      `json:"genderId"`
	PrimaryPhoneNumber string      `json:"primaryPhoneNumber"`
}

type ClientBodyAddress struct {
	Street          string `json:"street"`
	AddressLine1    string `json:"addressLine1"`
	AddressLine2    string `json:"addressLine2"`
	AddressLine3    string `json:"addressLine3"`
	City            string `json:"city"`
	StateProvinceId int    `json:"stateProvinceId"`
	CountryId       int    `json:"countryId"`
	PostalCode      int    `json:"postalCode"`
}
