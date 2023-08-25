/*
	NanoMon

	O que isso faz?

	- Eu tenho uma estação que faz um post no endereço /estacao/id
	- Eu tenho um servidor que escuta requisições post no endereço /estacao/:id
	- Eu tenho um cliente que faz long polling no endereço /estacao/stream
	- O servidor serve o cliente web.
*/

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Alfrederson/NanoIOT/pubsubber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type M map[string]interface{}

type Message struct {
	Time   time.Time `json:"time"`
	Device string    `json:"device"`
	Data   string    `json:"data"`
}

func (m *Message) ToJSON() string {
	jsonData, err := json.Marshal(m)
	if err != nil {
		// Handle the error, return an error string, or do something else
		return ""
	}
	return string(jsonData)
}

type Dashboard struct {
	LastMessages Stack
}

func Device(a *fiber.App, d *Dashboard) {
	// TODO: usar authorization header pra fazer autenticação dos dispositivos.

	a.Post("/dev/:id", func(c *fiber.Ctx) error {
		deviceId := c.Params("id")
		msg := Message{
			Time:   time.Now(),
			Device: deviceId,
			Data:   string(c.Body()),
		}

		d.LastMessages.Push(msg)

		message := msg.ToJSON()
		// publica assim: torradeira: Temperatura=10C Umidade=20% Coisa=X
		// recebe assim:
		//
		// {"time" : horário, "device" : id, "data" : aquilo que eu recebi}
		pubsubber.Publish("/dev/"+deviceId, message)
		pubsubber.Publish("/dev", message)

		return c.SendString("ok")
	})
	a.Get("/dev/:id", KeepAlive, func(c *fiber.Ctx) error {
		deviceId := c.Params("id")
		messageChannel := make(chan string, 1)

		_, _ = pubsubber.Subscribe("/dev/"+deviceId, messageChannel)
		return c.SendString(<-messageChannel)
	})

	a.Get("/dev/", KeepAlive, func(c *fiber.Ctx) error {
		log.Println("new listener")
		defer log.Println("fulfilled")
		messageChannel := make(chan string, 1)
		_, _ = pubsubber.Subscribe("/dev", messageChannel)
		return c.SendString(<-messageChannel)
	})
}

func WebClient(a *fiber.App, d *Dashboard) {
	a.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", M{
			"LastMessages": d.LastMessages.values,
		})
	})
}

func main() {

	dashboard := &Dashboard{
		LastMessages: Stack{
			values:   make([]interface{}, 0, 10),
			capacity: 10,
		},
	}

	engine := html.New("./web", ".html")
	engine.Reload(true)
	engine.AddFunc("time", func(t time.Time) string {
		j, err := json.Marshal(t)
		if err != nil {
			return "?"
		}
		r := string(j)
		return r[1 : len(r)-1]
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	Device(app, dashboard)
	WebClient(app, dashboard)

	app.Listen(":5000")
}
