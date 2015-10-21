package speech

import (
	"fmt"
	"log"
	"os"
)

const (
	speechURL   = "stream.watsonplatform.net/speech-to-text/api%s"
	sessionPATH = "/v1/sessions"
)

// Credentials is struct to save informations api
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Session  string `json:"session"`
}

// Setup !
func (c *Credentials) Setup() {
	usr := os.Getenv("SPEECH_USERNAME")
	psw := os.Getenv("SPEECH_PASSWORD")

	if usr != "" && psw != "" {
		c.Username = usr
		c.Password = psw
	} else {
		log.Fatal("Export SPEECH_(USERNAME/PASSWORD) environ vars")
	}
}

// MakeSessionURL Getting session url with username and password environments
func (c *Credentials) MakeSessionURL() string {
	url := fmt.Sprintf(speechURL, sessionPATH)
	sessionURL := fmt.Sprintf("https://%s:%s@%s", c.Username, c.Password, url)
	return sessionURL
}
