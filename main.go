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
	"os"
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

type Config struct {
	secret            string
	sgwBaseURL        string
	districtId        string
	serverPort        string
	couchbaseURL      string
	couchbaseReadsDB  string
	couchbaseWritesDB string
	couchbaseUser     string
	couchbasePass     string
}

//global envs map

func main() {

	//populate config from environment variables
	config := Config{
		secret:            os.Getenv("SECRET"),
		sgwBaseURL:        os.Getenv("SGW_BASE_URL"),
		districtId:        os.Getenv("DISTRICT_ID"),
		serverPort:        os.Getenv("SERVER_PORT"),
		couchbaseURL:      os.Getenv("COUCHBASE_URL"),
		couchbaseReadsDB:  os.Getenv("COUCHBASE_READS_DB"),
		couchbaseWritesDB: os.Getenv("COUCHBASE_WRITES_DB"),
		couchbaseUser:     os.Getenv("COUCHBASE_USER"),
		couchbasePass:     os.Getenv("COUCHBASE_PASS"),
	}

	//check if all config values are set
	if config.secret == "" || config.sgwBaseURL == "" || config.serverPort == "" || config.couchbaseURL == "" || config.couchbaseReadsDB == "" || config.couchbaseWritesDB == "" || config.couchbaseUser == "" || config.couchbasePass == "" {
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

		config = Config{
			secret:            envs["SECRET"],
			sgwBaseURL:        envs["SGW_BASE_URL"],
			districtId:        envs["DISTRICT_ID"],
			serverPort:        envs["SERVER_PORT"],
			couchbaseURL:      envs["COUCHBASE_URL"],
			couchbaseReadsDB:  envs["CB_DB"],
			couchbaseWritesDB: envs["CB_WRITES_DB"],
			couchbaseUser:     envs["CB_USER"],
			couchbasePass:     envs["CB_PASS"],
		}

	}

	couchbaseService := data.NewService(config.couchbaseURL, config.couchbaseReadsDB, config.couchbaseWritesDB, config.couchbaseUser, config.couchbasePass)

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
				Secret:     config.secret,
				SGWBaseURL: config.sgwBaseURL,
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
	log.Println("Server started on port " + config.serverPort)
	log.Fatal(http.ListenAndServe(":"+config.serverPort, nil))
}
