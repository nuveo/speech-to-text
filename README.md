# Speech to Text
IBM Bluemix, the Speech to Text service converts the human voice into the written word.
### README WIP
### Install

`go get github.com/poorny/speech-to-text`

Export to env `SPEECH_USERNAME` and `SPEECH_PASSWORD`.

_credentials getting in dashboard_


### Usage

##### bin
```
cd $GOPATH/src/github.com/poorny/speech-to-text/speech
go build
./speech -input /path/to/audio/file
```
##### Code

```go
package main

import (
  "fmt"
  "github.com/poorny/speech-to-text"
)
func main() {
  c := speech.Credentials{}
  c.Setup()

  url := c.MakeSessionURL()
  sess, err := speech.GetSession(url)
  if err != nil {
    return
  }

  status, err := sess.GetRecognize()
  if err != nil {
    return
  }

  if status.State != "initialized" {
    log.Println("Not ready yet!")
    return
  }

  text, err := sess.SendAudio(*inputFile)
  if err != nil {
    return
  }

  fmt.Println(text)
  sess.DeleteSession()
}

```
