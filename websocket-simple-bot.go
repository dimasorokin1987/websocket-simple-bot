package main

import (
  "fmt"
  "os"
  "net/http"
  "log"
  //"encoding/json"
  "golang.org/x/net/websocket"
  wit "github.com/christianrondeau/go-wit"
)

const (
  directory = "./web"
  confidenceThreshold = 0.9
)

type T struct {
  Txt string `json:"text"`
}

var (
  witClient *wit.Client
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello World!")
}

func WebsocketServer(ws *websocket.Conn) {
  for {
    data := T{}
    err := websocket.JSON.Receive(ws, &data) 
    if err != nil {
      log.Fatalln("error receiving json")
    }
    result, err := witClient.Message(data.Txt)
    if err != nil {
      log.Fatalln("error wit processing message")
    }
    //log.Printf("%v\n", result)

    var (
      topEntity    wit.MessageEntity
      topEntityKey string
    )

    for key, entityList := range result.Entities {
      for _, entity := range entityList {
        if entity.Confidence > confidenceThreshold && entity.Confidence > topEntity.Confidence {
            topEntity = entity
            topEntityKey = key
        }
      }
    }

    switch topEntityKey {
      case "greetings":
        data.Txt = "Hello, user! How can I help you?"
      case "wolfram_search_query":
        // TODO
    }

    websocket.JSON.Send(ws, data)
  }
}

func main() {
  port := os.Getenv("PORT")
  witAiAccessToken := os.Getenv("WIT_AI_ACCESS_TOKEN")
  witClient = wit.NewClient(witAiAccessToken)

  mux := http.NewServeMux()
  mux.Handle("/", http.FileServer(http.Dir(directory)))
  mux.Handle("/test", http.HandlerFunc(indexHandler))
  mux.Handle("/ws", websocket.Handler(WebsocketServer))
  s := http.Server{Addr: ":" + port, Handler: mux}
  err := s.ListenAndServe()
  if err != nil {
    log.Fatalln("ListenAndServe: " + err.Error())
  }
}
