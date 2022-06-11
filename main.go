package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gempir/go-twitch-irc/v2"
	"github.com/joho/godotenv"
)

type SRBody struct {
	Name string `json:"name"`
	Bot  bool   `json:"bot"`
}

func main() {
	lg := log.Default()
	lg.SetPrefix("DEBUG ")
	if err := godotenv.Load(); err != nil {
		log.Println(".env missing")
	}

	channel := os.Getenv("CHANNEL_NAME")
	if len(channel) == 0 {
		lg.Panicln("channel missing")
	}
	username := os.Getenv("BOT_USERNAME")
	token := os.Getenv("OAUTH_TOKEN")

	var client *twitch.Client
	if len(username) == 0 || len(token) == 0 {
		client = twitch.NewAnonymousClient()
	} else {
		client = twitch.NewClient(username, token)
	}

	client.Join(channel)
	client.OnConnect(func() {
		lg.Println("connected")
	})
	client.OnReconnectMessage(func(message twitch.ReconnectMessage) {
		lg.Println("reconnected")
	})
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		exp := "^(點:|點：).*"
		r, _ := regexp.Compile(exp)
		m := r.FindStringSubmatch(message.Message)
		if len(m) == 0 || len(m) > 2 {
			return
		}
		c := m[1]
		sr := strings.Replace(message.Message, c, "", 1)
		sr = strings.Trim(sr, " ")
		lg.Println(sr)
		rq := SRBody{
			Name: sr,
			Bot:  true,
		}
		bt, err := json.Marshal(rq)
		if err != nil {
			lg.Panicln(err)
		}
		_, err = http.Post("http://localhost:4001/api/request", "application/json", bytes.NewBuffer(bt))
		if err != nil {
			lg.Println("failed")
		} else {
			lg.Println("success")
			client.Say(message.Channel, fmt.Sprintf("%s 成功點了%s", message.User.DisplayName, sr))
		}
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}

	defer client.Disconnect()
}
