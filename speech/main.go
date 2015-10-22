package main

import (
	"flag"
	"fmt"

	"github.com/poorny/speech-to-text"
)

var (
	inputFile = flag.String("input", "", "File to convert into text")
)

func main() {
	flag.Parse()

	if *inputFile != "" {
		c := speech.Credentials{}
		c.Setup()

		url := c.MakeSessionURL()
		sess, err := speech.GetSession(url)
		if err != nil {
			return
		}

		sess.GetRecognize()

		text, err := sess.SendAudio(*inputFile)
		if err != nil {
			return
		}

		fmt.Println(text)
		sess.DeleteSession()
	}
}
