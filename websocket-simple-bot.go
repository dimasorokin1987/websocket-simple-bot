package main

import (
  "fmt"
  "os"
  "io"
  "net/http"
  "log"
  "golang.org/x/net/websocket"
)

const (
  directory = "./web"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
//   if origin := r.Header.Get("Origin"); origin != "" {
//     w.Header().Set("Access-Control-Allow-Origin", origin)
//     w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
//     w.Header().Set("Access-Control-Allow-Headers",
//         "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
//   }
//   // Stop here if its Preflighted OPTIONS request
//   if r.Method == "OPTIONS" {
//     return
//   }
  fmt.Fprintf(w, "Hello World!")
}

// Echo the data received on the WebSocket. 
func EchoServer(ws *websocket.Conn) {
  io.Copy(ws, ws)
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
