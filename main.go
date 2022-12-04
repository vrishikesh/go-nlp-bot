package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Krognol/go-wolfram"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

func printCommandEvents(analyticsChan <-chan *slacker.CommandEvent) {
	for event := range analyticsChan {
		log.Println("Command Event")
		log.Println(event.Timestamp)
		log.Println(event.Command)
		log.Println(event.Parameters)
		log.Println(event.Event)
		log.Println("=============")
	}
}

func main() {
	godotenv.Load(".env")

	bot := slacker.NewClient(
		os.Getenv("SLACK_BOT_TOKEN"),
		os.Getenv("SLACK_APP_TOKEN"),
	)
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_TOKEN")}

	re := regexp.MustCompile(`<.*>`)

	go printCommandEvents(bot.CommandEvents())

	bot.Command("<message>", &slacker.CommandDefinition{
		Description: "sends any question to wolfram",
		Examples:    []string{"who is the president of india"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			log.Println("raw -", request.Param("message"))

			query := re.ReplaceAllString(request.Param("message"), "")
			query = strings.TrimSpace(query)
			log.Println("query -", query)

			msg, err := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			if err != nil {
				response.Reply(err.Error())
				return
			}
			log.Println("msg -", msg)

			data, _ := json.MarshalIndent(msg, "", "  ")
			rough := string(data[:])
			log.Println("rough -", rough)

			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			log.Println("value -", value)

			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				response.Reply(err.Error())
				return
			}

			response.Reply(res)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
