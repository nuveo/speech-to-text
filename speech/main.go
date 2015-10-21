package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/poorny/speech"
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
		sess, f := speech.GetSession(url)
		if f == false {
			return
		}

		sess.GetRecognize()

		text, f := sess.SendAudio(*inputFile)
		if f == false {
			return
		}
		fmt.Println(text)
		sess.DeleteSession()
	}
	log.Println("Send a file with -input flag")
}
