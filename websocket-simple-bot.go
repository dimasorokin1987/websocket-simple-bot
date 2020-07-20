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

type SlackMessage struct {
  Type string `json:"type"`
  SubType string `json:"subtype"`
  Text string `json:"text"`
  TimeStamp string `json:"ts"`
  UserName string `json:"username"`
  BotId string `json:"bot_id"`
}

type SlackMessageResult struct {
  Ok bool `json:"ok"`
  Channel string `json:"channel"`
  TimeStamp string `json:"ts"`
  Message SlackMessage `json:"message"`
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
      log.Println("error receiving json")
      continue
    }
    result, err := witClient.Message(data.Txt)
    if err != nil {
      log.Println("error wit processing message")
      continue
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

    //data.Txt = topEntityKey
    switch topEntityKey {
      case "greetings":
        data.Txt = "Hello, user! How can I help you?"
      case "wolfram_search_query":
        res, err := wolframClient.GetSpokentAnswerQuery(topEntity.Value.(string), wolfram.Metric, 1000)
        if err == nil {
          data.Txt = res
        } else {
          log.Println("wolfram client: " + err.Error())
          continue
        }
      case "intent":
        data.Txt = "Enter your message:"
        websocket.JSON.Send(ws, data)
        err := websocket.JSON.Receive(ws, &data) 
        if err != nil {
          log.Println("error user message receiving json")
          continue
        }
        requestBody, err := json.Marshal(map[string]string{
          "channel": "#general",
          "text": data.Txt,
        })
        if err != nil {
          log.Println(err)
          continue
        }
        timeout := time.Duration(5*time.Second)
        client := http.Client{
          Timeout: timeout,
        }
        request, err := http.NewRequest("POST", slackUrl, bytes.NewBuffer(requestBody))
        if err != nil {
          log.Println(err)
          continue
        }
        request.Header.Set("Content-Type","application/json;charset=utf-8")
        request.Header.Set("Authorization","Bearer "+slackSecretKey)
        resp, err := client.Do(request)
        if err != nil {
          log.Println(err)
          continue
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
          log.Println(err)
          continue
        }
        //data.Txt = string(body)
        slackRes := SlackMessageResult{}
        json.Unmarshal([]byte(body), &slackRes)
        log.Println(slackRes)
        if slackRes.Ok && slackRes.Message.Text == data.Txt {
          data.Txt = "Message was sended success"
        } else {
          data.Txt = "Fail to send slack message"
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
