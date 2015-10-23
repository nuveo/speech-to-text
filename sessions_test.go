package speech

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// GetSession test
// SendAudio test
// GetRecognize test
// ObserverResult test
// DeleteSession test

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func makeMockSession(url string) SessionRsp {
	mockSessionID := GetMD5Hash("mock")

	sessionStrc := SessionRsp{
		SessionID:     mockSessionID,
		NewSessionURI: fmt.Sprintf("http://%s/speech-to-text/api/v1/sessions/%s", url, mockSessionID),
		Recognize:     fmt.Sprintf("http://%s/speech-to-text/api/v1/sessions/%s/observe_result", url, mockSessionID),
		ObserveResult: fmt.Sprintf("http://%s/speech-to-text/api/v1/sessions/%s/recognize", url, mockSessionID),
	}
	return sessionStrc
}

func HgetSession(w http.ResponseWriter, r *http.Request) {
	mockSession := makeMockSession(r.Host)
	response, err := json.Marshal(mockSession)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(response)
}

func HsendAudio(w http.ResponseWriter, r *http.Request) {}

func HgetRecognize(w http.ResponseWriter, r *http.Request) {
	mockSession := makeMockSession(r.Host)
	var s = struct {
		State         string
		Model         string
		Recognize     string
		ObserveResult string
	}{
		"initialized", "pt-BR_BroadbandModel",
		mockSession.Recognize, mockSession.ObserveResult,
	}
	response, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(response)
}

func HobserveResult(w http.ResponseWriter, r *http.Request) {}

func HdeleteSession(w http.ResponseWriter, r *http.Request) {}

func handlers() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/getSession/", HgetSession)
	r.HandleFunc("/sendAudio/", HsendAudio)
	r.HandleFunc("/getRecognize/", HgetRecognize)
	r.HandleFunc("/observeResult/", HobserveResult)
	r.HandleFunc("/deleteSession/", HdeleteSession)

	return r
}

func setupTestHandlers() *httptest.Server {
	setupTest()

	hand := handlers()
	serve := httptest.NewServer(hand)
	return serve
}

func TestGetSession(t *testing.T) {
	server := setupTestHandlers()

	c := Credentials{}
	c.Setup()

	urlToSession := fmt.Sprintf("%s%s", server.URL, "/getSession/")
	sess, err := GetSession(urlToSession)
	if err != nil {
		log.Println("Error:", err)
	}
	if sess.NewSessionURI == "" {
		t.Error("NewSessionURI empty!")
	}
	if sess.ObserveResult == "" {
		t.Error("ObserveResult empty!")
	}
	if sess.Recognize == "" {
		t.Error("Recognize empty!")
	}
	if sess.SessionID == "" {
		t.Error("SessionID empty!")
	}
}

func TestGetRecognize(t *testing.T) {
	server := setupTestHandlers()

	c := Credentials{}
	c.Setup()

	urlToSession := fmt.Sprintf("%s%s", server.URL, "/getSession/")
	sess, err := GetSession(urlToSession)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	// Mock URL
	urlToTest := fmt.Sprintf("%s%s", server.URL, "/getRecognize/")
	sess.Recognize = urlToTest

	status, err := sess.GetRecognize()
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	if status.State != "initialized" {
		t.Error("State should be `initialized`")
	}

}
