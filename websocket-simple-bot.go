package main

import (
  "fmt"
  "os"
  "net/http"
  "log"
  "golang.org/x/net/websocket"
  "github.com/christianrondeau/go-wit"
)

const (
  directory = "./web"
)

type T struct {
  Txt string `json:"text"`
}

var (
  witClient *Client
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
    log.Println(result.Entities["intent"].value)
    data.Txt = result.Entities["intent"].value
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
