package main

import (
  "fmt"
  "os"
  "net/http"
  "log"
  "encoding/json"
  "golang.org/x/net/websocket"
  wit "github.com/christianrondeau/go-wit"
)

const (
  directory = "./web"
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
    //result, err := witClient.Parse(&wit.MessageRequest{
    //  Query: "hello",
    //})
    if err != nil {
      log.Fatalln("error wit processing message")
    }
    //log.Println(result.Entities["intent"].value)
    log.Println(result)
    //data, _ := json.MarshalIndent(result, "", "    ")
    //log.Println(string(data[:]))

    data.Txt = result
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
