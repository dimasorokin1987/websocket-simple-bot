package main

import (
  "fmt"
  "os"
  "net/http"
  "log"
  "encoding/json"
  "io/ioutil"
  "bytes"
  "time"

  "golang.org/x/net/websocket"
  wit "github.com/christianrondeau/go-wit"
  wolfram "github.com/Krognol/go-wolfram"
)

const (
  directory = "./web"
  confidenceThreshold = 0.9
  slackUrl = "https://slack.com/api/chat.postMessage"

)

type T struct {
  Txt string `json:"text"`
}

var (
  witClient *wit.Client
  wolframClient *wolfram.Client
  slackSecretKey string
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
      case "slackMessage":
        data.Txt = "Enter your message:"
        websocket.JSON.Send(ws, data)
        err := websocket.JSON.Receive(ws, &data) 
        if err != nil {
          log.Println("error user message receiving json")
        }
        requestBody, err := json.Marshal(map[string]string{
          "channel": "#general",
          "text": data.Txt,
        })
        if err != nil {
          log.Println(err)
        }
        timeout := time.Duration(5*time.Second)
        client := http.Client{
          Timeout: timeout,
        }
        request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
        if err != nil {
          log.Println(err)
        }
        request.Header.Set("Content-Type","application/json;charset=utf-8")
        request.Header.Set("Authorization","Bearer "+slackSecretKey)
        resp, err := client.Do(request)
        if err != nil {
          log.Println(err)
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
          log.Println(err)
        }
        data.Txt = string(body)
        //data.Txt = "Message was sended success"
        websocket.JSON.Send(ws, data)
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
  slackSecretKey = os.Getenv("SLACK_SECRET_KEY")

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
