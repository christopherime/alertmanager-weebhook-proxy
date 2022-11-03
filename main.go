package main

// alertmanager-weebhook-proxy
// Description: A simple proxy for alertmanager webhooks

// TODO: Implement auth method switch
// TODO: Implement auth credentials decryption (and remove the secret from the config file)

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// DATA STRUCTURES

type ConfigObject struct {
	Server struct {
		Port     string `yaml:"port"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"server"`
	Targets []struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"targets"`
}

type AlertObject struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   []struct {
		Status string `json:"status"`
		Labels struct {
			Alertname string `json:"alertname"`
			Service   string `json:"service"`
			Severity  string `json:"severity"`
		} `json:"labels"`
		Annotations struct {
			Summary string `json:"summary"`
		} `json:"annotations"`
		StartsAt     string    `json:"startsAt"`
		EndsAt       time.Time `json:"endsAt"`
		GeneratorURL string    `json:"generatorURL"`
		Fingerprint  string    `json:"fingerprint"`
	} `json:"alerts"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	CommonLabels struct {
		Alertname string `json:"alertname"`
		Service   string `json:"service"`
		Severity  string `json:"severity"`
	} `json:"commonLabels"`
	CommonAnnotations struct {
		Summary string `json:"summary"`
	} `json:"commonAnnotations"`
	ExternalURL string `json:"externalURL"`
	Version     string `json:"version"`
	GroupKey    string `json:"groupKey"`
}

// loadConfig function read the config store in config.yml
// and return a ConfigObject struct
func loadConfig() ConfigObject {
	var config ConfigObject

	// Read ./config.yml
	ymlConfig, err := os.ReadFile("./config.yml")
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal config.yml
	err = yaml.Unmarshal(ymlConfig, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func main() {

	// Load config
	config := loadConfig()
	_ = config

	log.Println("Starting alertmanager-weebhook-proxy")
	log.Println("Listening on port " + config.Server.Port + " for " + config.Server.Endpoint)

	http.HandleFunc(config.Server.Endpoint, func(w http.ResponseWriter, r *http.Request) {

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Unmarshal request body
		var alert AlertObject
		err = json.Unmarshal(body, &alert)
		if err != nil {
			log.Fatal(err)
		}

		// Log alert
		log.Println(alert)

		// Send alert to targets
		for _, target := range config.Targets {

			// Create request
			req, err := http.NewRequest("POST", target.URL, r.Body)
			if err != nil {
				log.Fatal(err)
			}

			// Set headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", target.Token)

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			// Log response
			log.Println(resp)

		}

	})

	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))

}
