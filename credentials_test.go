package speech

import (
	"fmt"
	"os"
	"testing"
)

// Setup test
// MakeSessionURL test

// setupTest set fake environ vars
func setupTest() {
	os.Setenv("SPEECH_USERNAME", "mock_username")
	os.Setenv("SPEECH_PASSWORD", "mock_password")
}

func TestCredentialsSetup(t *testing.T) {
	setupTest()
	c := new(Credentials)
	c.Setup()

	if c.Username == "" {
		t.Errorf("SPEECH_USERNAME environ not avaliable")
	}

	if c.Password == "" {
		t.Errorf("SPEECH_PASSWORD environ not avaliable")
	}
}

func TestCredentialsMakeSessionURL(t *testing.T) {
	setupTest()
	c := new(Credentials)
	c.Setup()

	url := c.MakeSessionURL()
	sessionURL := fmt.Sprintf(speechURL, sessionPATH)
	expected := fmt.Sprintf("https://%s:%s@%s", "mock_username", "mock_password", sessionURL)

	if url != expected {
		t.Errorf("Should be %s got %s", expected, url)
	}
}
