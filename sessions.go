package speech

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
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

// RecognizeStatus Get status from /recgonize api.
// checks that Speech-to-text api is available for new recognition
type RecognizeStatus struct {
	Session RecognizeBody `json:"session"`
}

// RecognizeBody body response
type RecognizeBody struct {
	State         string `json:"state"`
	Model         string `json:"model"`
	Recognize     string `json:"recognize"`
	ObserveResult string `json:"observe_result"`
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
func GetSession(sessionURL string) (SessionRsp, error) {
	log.Println("Getting session")
	jsonStr := []byte(`{}`)
	modelURL := fmt.Sprintf("%s?model=%s", sessionURL, "pt-BR_BroadbandModel")

	req, err := http.NewRequest("POST", modelURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println("Creating new Request", err)
		return SessionRsp{}, err
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
		return SessionRsp{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Reading body", err)
		return SessionRsp{}, err
	}

	if resp.StatusCode != 201 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return SessionRsp{}, err
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		errF := fmt.Sprintf("%s - %s", errorStc.Error, errorStc.CodeDescription)
		return SessionRsp{}, errors.New(errF)
	}

	var sessionRsp SessionRsp
	err = json.Unmarshal(body, &sessionRsp)
	if err != nil {
		log.Println("Unmarshal body", err)
		return SessionRsp{}, err
	}
	sessionRsp.CJar = cookiesJar
	log.Println("Getting session - Done")
	return sessionRsp, nil
}

// SendAudio blah
func (s *SessionRsp) SendAudio(pathAudio string) (string, error) {
	log.Println("Seding audio")
	path, err := ConvertToWav(pathAudio)
	if err != nil {
		log.Println("Error on Convert", err)
		return "", err
	}

	wav, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s?continuous=true", s.Recognize), wav)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "audio/wav")
	req.Header.Set("Transfer-Encoding", "Chunked")

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return "", err
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		errF := fmt.Sprintf("%s - %s", errorStc.Error, errorStc.CodeDescription)
		return "", errors.New(errF)
	}

	var response RecognizeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Unmarshall body", err)
		return "", err
	}

	final := []string{}

	for _, resp := range response.Results {
		if resp.Final {
			for _, alt := range resp.Alternatives {
				if _, ok := alt["confidence"]; ok {
					final = append(final, alt["transcript"].(string))
				}
			}
		}
	}
	log.Println("Send Audio - Done")
	if len(final) == 0 {
		return "", errors.New("nothing was recognized")
	}
	return strings.Join(final, " "), err
}

// GetRecognize -<
func (s *SessionRsp) GetRecognize() (RecognizeStatus, error) {
	log.Println("Get Recognize Status")
	req, err := http.NewRequest("GET", s.Recognize, nil)
	if err != nil {
		log.Println(err)
		return RecognizeStatus{}, err
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return RecognizeStatus{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return RecognizeStatus{}, err
	}
	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return RecognizeStatus{}, err
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		errF := fmt.Sprintf("%s - %s", errorStc.Error, errorStc.CodeDescription)
		return RecognizeStatus{}, errors.New(errF)
	}

	var recognizeSts RecognizeStatus
	err = json.Unmarshal(body, &recognizeSts)
	if err != nil {
		log.Println("Unmarshal body", err)
		return RecognizeStatus{}, err
	}
	log.Println("Get Recognize Status - Done")
	return recognizeSts, nil
}

// ObserverResult <-
func (s *SessionRsp) ObserverResult() error {
	log.Println("Observe Result url")
	url := makeURLCredentials(s.ObserveResult)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.StatusCode != 200 {
		var errorStc ErrorResponse
		err = json.Unmarshal(body, &errorStc)
		if err != nil {
			log.Println("Unmarshal error body", err)
			return err
		}
		log.Println(errorStc.Error, errorStc.CodeDescription)
		errF := fmt.Sprintf("%s - %s", errorStc.Error, errorStc.CodeDescription)
		return errors.New(errF)
	}

	log.Println("Observe Result - Done")
	return nil
}

// DeleteSession remove session
func (s *SessionRsp) DeleteSession() error {
	log.Println("Delete SESSION")

	req, err := http.NewRequest("DELETE", s.NewSessionURI, nil)
	if err != nil {
		log.Println("Make DELETE request", err)
		return err
	}

	client := http.Client{
		Jar: s.CJar,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send Delete request", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	if resp.StatusCode == 204 {
		log.Println("Session closed!")
		return nil
	}

	var errorStc ErrorResponse
	err = json.Unmarshal(body, &errorStc)
	if err != nil {
		log.Println("Unmarshal error body", err)
		return err
	}
	log.Println(errorStc.Error, errorStc.CodeDescription)
	errF := fmt.Sprintf("%s - %s", errorStc.Error, errorStc.CodeDescription)
	return errors.New(errF)
}
