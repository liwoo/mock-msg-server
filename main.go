package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"mock-server/auth"
	"mock-server/data"
	"net/http"
)

type LoginRequestDTO struct {
	JWT string `json:"jwt"`
}

type ApiRequestDTO struct {
	Id string `json:"id"`
}

func writeError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(err)
}

//global envs map

func main() {

	envFiles := []string{".env", "../.env"}
	availableEnvFile := ""
	var envs map[string]string

	for _, envFile := range envFiles {
		//check if file exists
		if _, err := ioutil.ReadFile(envFile); err == nil {
			availableEnvFile = envFile
			break
		}
	}

	if availableEnvFile == "" {
		log.Fatal("No env file found")
	}

	envs, err := godotenv.Read(availableEnvFile)

	if err != nil {
		log.Fatal(err)
	}

	//to be read from a config file
	var (
		secret            = envs["KC_PUBLIC_KEY"]
		serverPort        = envs["SERVER_PORT"]
		sgwBaseURL        = envs["SGW_BASE_URL"]
		couchbaseURL      = envs["CB_URL"]
		couchbaseReadsDB  = envs["CB_DB"]
		couchbaseWritesDB = envs["CB_WRITES_DB"]
		couchbaseUser     = envs["CB_USER"]
		couchbasePass     = envs["CB_PASS"]
	)

	couchbaseService := data.NewService(couchbaseURL, couchbaseReadsDB, couchbaseWritesDB, couchbaseUser, couchbasePass)

	//http server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Welcome to Mobile Sync Gateway Mock Server"))
		if err != nil {
			return
		}
	})

	http.HandleFunc("/api/v3/api-requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
			}

			var requestDTO ApiRequestDTO
			err = json.Unmarshal(body, &requestDTO)

			if err != nil {
				writeError(w, err, http.StatusBadRequest)
			}

			err = couchbaseService.ProcessApiRequest(requestDTO.Id)

			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusCreated)
			message := fmt.Sprintf("Successfully created API request with id: %s", requestDTO.Id)
			_, err = w.Write([]byte(message))
		}
	})

	http.HandleFunc("/api/v3/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			//get the body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
				return
			}

			//unmarshal the body
			var loginRequestDto LoginRequestDTO

			err = json.Unmarshal(body, &loginRequestDto)

			if err != nil {
				writeError(w, err, http.StatusBadRequest)
				return
			}

			// Parse the token
			authService := auth.Service{
				JWT:        loginRequestDto.JWT,
				Secret:     secret,
				SGWBaseURL: sgwBaseURL,
			}

			claims, err := authService.RetrieveClaims()

			if err != nil {
				writeError(w, err, http.StatusUnauthorized)
				return
			}

			created, err := authService.CreateSGWUser(&claims)

			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
				return
			}

			//json marshall created
			createdJson, err := json.Marshal(created)

			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
				return
			}

			go func() {
				docs, err := couchbaseService.PublishDocs(created.AdminChannels)

				if err != nil {
					log.Println(err)
				}

				log.Println("Published", docs, "documents...")
			}()

			if err != nil {
				log.Println(err)
			}

			_, err = w.Write([]byte(fmt.Sprintf("%s", createdJson)))

			if err != nil {
				writeError(w, err, http.StatusInternalServerError)
				return
			}
		}
	})

	//run the server with a message
	log.Println("Server started on port " + serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
