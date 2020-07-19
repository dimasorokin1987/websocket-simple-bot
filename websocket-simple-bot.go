package main

import (
  "fmt"
  "os"
  //"io"
  "net/http"
  "log"
  "golang.org/x/net/websocket"
)

const (
  directory = "./web"
)

type T struct {
  Msg string
  Count int
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello World!")
}

// Echo the data received on the WebSocket. 
func EchoServer(ws *websocket.Conn) {
  //go func()
  //io.Copy(ws, ws)
  data := T{}
  err := websocket.JSON.Receive(ws, &data) 
  if err != nil {
    log.Fatalln("error receiving json")
  }
  err := websocket.JSON.Send(ws, data)
  if err != nil {
    log.Fatalln("error sending json")
  }
}

func main() {
  http.Handle("/", http.FileServer(http.Dir(directory)))
  http.Handle("/ws", websocket.Handler(EchoServer))
  // err := http.ListenAndServe(":12345", nil)
  http.Handle("/test", http.HandlerFunc(indexHandler))
  err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
  if err != nil {
    log.Fatalln("ListenAndServe: " + err.Error())
  }
}
