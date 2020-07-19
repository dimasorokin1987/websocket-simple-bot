package main

import (
  "fmt"
  "os"
  "net/http"
  "log"
  //"encoding/json"
  "golang.org/x/net/websocket"
  wit "github.com/christianrondeau/go-wit"
  wolfram "github.com/Krognol/go-wolfram"
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
  wolframClient *wolfram.Client
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
        res, err := wolframClient.GetSpokentAnswerQuery(topEntity.Value.(string), wolfram.Metric, 1000)
        if err == nil {
          data.Txt = res
        } else {
          log.Println("wolfram client: " + err.Error())
        }
      default:
        data.Txt = "¯\\_(o_o)_/¯"
    }

    websocket.JSON.Send(ws, data)
  }
}

func main() {
  port := os.Getenv("PORT")
  witAiAccessToken := os.Getenv("WIT_AI_ACCESS_TOKEN")
  wolframAppId := os.Getenv("WOLFRAM_APP_ID")

  witClient = wit.NewClient(witAiAccessToken)
  wolframClient = &wolfram.Client{AppID: wolframAppId}

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
