package speech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
)

// Result strct
type Result struct {
	Final        bool                     `json:"final"`
	Alternatives []map[string]interface{} `json:"alternatives"`
}

// RecognizeResponse strct
type RecognizeResponse struct {
	ResultIndex int      `json:"result_index"`
	Results     []Result `json:"results"`
}

// ErrorResponse struct of errors
type ErrorResponse struct {
	Error           string `json:"error"`
	CodeDescription string `json:"code_description"`
}

// SessionRsp is struct to save session informations from request
type SessionRsp struct {
	SessionID     string `json:"session_id"`
	NewSessionURI string `json:"new_session_uri"`
	Recognize     string `json:"recognize"`
	ObserveResult string `json:"observe_result"`
	CJar          *cookiejar.Jar
}

// makeURLCredentials -<
func makeURLCredentials(url string) string {
	usr := os.Getenv("SPEECH_USERNAME")
	psw := os.Getenv("SPEECH_PASSWORD")

	if usr != "" && psw != "" {
		nURL := fmt.Sprintf("https://%s:%s@%s", usr, psw, url[8:])
		return nURL
	}

	log.Fatal("Export SPEECH_(USERNAME/PASSWORD) environ vars")
	return ""
}

// GetSession <-
func GetSession(sessionURL string) (SessionRsp, bool) {
	log.Println("Getting session")
	jsonStr := []byte(`{}`)
	modelURL := fmt.Sprintf("%s?model=%s", sessionURL, "pt-BR_BroadbandModel")

	req, err := http.NewRequest("POST", modelURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("Creating new Request", err)
		return SessionRsp{}, false
	}
	req.Header.Set("Content-Type", "application/json")
	cookiesJar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{
		Jar: cookiesJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send request", err)
		return SessionRsp{}, false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Reading body", err)
		return SessionRsp{}, false
	}

	if resp.StatusCode != 201 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return SessionRsp{}, false
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		return SessionRsp{}, false
	}

	var sessionRsp SessionRsp
	err = json.Unmarshal(body, &sessionRsp)
	if err != nil {
		log.Println("Unmarshal body", err)
		return SessionRsp{}, false
	}
	sessionRsp.CJar = cookiesJar
	log.Println("Getting session - Done")
	return sessionRsp, true
}

// SendAudio blah
func (s *SessionRsp) SendAudio(pathAudio string) (string, bool) {
	log.Println("Seding audio")
	path, err := ConvertToWav(pathAudio)
	if err != nil {
		log.Println("Error on Convert", err)
		return "", false
	}

	wav, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return "", false
	}
	req, err := http.NewRequest("POST", s.Recognize, wav)
	if err != nil {
		log.Println(err)
		return "", false
	}
	req.Header.Set("Content-Type", "audio/wav")
	//req.Header.Set("Transfer-Encoding", "Chunked")

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", false
	}

	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return "", false
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		return "", false
	}

	var response RecognizeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Unmarshall body", err)
		return "", false
	}

	for _, resp := range response.Results {
		if resp.Final {
			for _, alt := range resp.Alternatives {
				if _, ok := alt["confidence"]; ok {
					log.Println("Send Audio - Done")
					return alt["transcript"].(string), true
				}
			}
		}
	}
	return "", false
}

// GetRecognize -<
func (s *SessionRsp) GetRecognize() bool {
	log.Println("Get Recognize Status")
	req, err := http.NewRequest("GET", s.Recognize, nil)
	if err != nil {
		log.Println(err)
		return false
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return false
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		return false
	}
	log.Println("Get Recognize Status - Done")
	return true
}

// ObserverResult <-
func (s *SessionRsp) ObserverResult() bool {
	log.Println("Observe Result url")
	url := makeURLCredentials(s.ObserveResult)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return false
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return false
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		return false
	}

	log.Println("Observe Result - Done")
	return true
}

// DeleteSession remove session
func (s *SessionRsp) DeleteSession() bool {
	log.Println("Delete SESSION")
	url := makeURLCredentials(s.NewSessionURI)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Println("Make DELETE request", err)
		return false
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send Delete request", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if resp.StatusCode == 204 {
		log.Println("Session closed!")
		return true
	}

	var errorStc ErrorResponse
	err = json.Unmarshal(body, &errorStc)
	if err != nil {
		log.Println("Unmarshal error body", err)
		return false
	}
	log.Println(errorStc.Error, errorStc.CodeDescription)
	return false
}
