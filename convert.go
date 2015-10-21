package speech

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// ConvertToWav convert files mp3 to Wav
func ConvertToWav(path string) (string, error) {
	log.Println("Converting")
	f, err := os.Open(path)
	if err != nil {
		return "", errors.New(err.Error())
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.New(err.Error())
	}

	response := http.DetectContentType(body)
	if response == "audio/wave" {
		log.Println("file is audio/wav type")
		return path, nil
	}

	nameFile, err := newUUID()
	if err != nil {
		return "", errors.New(err.Error())
	}

	tmpDir := fmt.Sprintf("/tmp/%s.wav", nameFile)
	cmd := fmt.Sprintf("avconv -i %s %s", path, tmpDir)

	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println(err, cmd)
	}
	log.Println("Done convert!")
	return tmpDir, err
}
