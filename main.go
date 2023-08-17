/*
	NanoIOT

	O que isso faz?

	- Eu tenho uma estação que faz um post no endereço /estacao/id
	- Eu tenho um servidor que escuta requisições post no endereço /estacao/:id
	- Eu tenho um cliente que faz long polling no endereço /estacao/stream
	- O servidor serve o cliente web.

*/

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Alfrederson/NanoIOT/pubsubber"
	"github.com/gofiber/fiber/v2"
)

func SubHandler(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	topic := c.Params("topic")
	client := c.Params("clientid")
	// o parâmetro clientid serve literalmente parada.
	// não sei por que, mas sem ele, se eu tenho 2 subscribers no mesmo tópico ao mesmo tempo,
	// o segundo leva uns 15 segundos pra começar a escutar.
	// tem algum motivo interno pra isso acontecer porque é a mesma coisa com o gin.
	log.Println(client, ": subscreveu a  ", topic)
	messageChannel := make(chan string, 1)
	// zumbi, mas quando estava com o gin isso servia pra poder remover o canal
	// quando o cliente desconectava.
	_, _ = pubsubber.Subscribe(topic, messageChannel)
	return c.SendString(<-messageChannel)
}

func PubHandler(c *fiber.Ctx) error {
	topic := c.Params("topic")
	message := c.Body()
	pubsubber.Publish(topic, string(message))
	return c.SendString("mensagem publicada")
}

func PubHandlerGet(c *fiber.Ctx) error {
	type MessageQuery struct {
		Message string
	}

	topic := c.Params("topic")
	message := MessageQuery{}
	if err := c.QueryParser(&message); err != nil {
		return c.Status(503).SendString("")
	}
	pubsubber.Publish(topic, message.Message)
	log.Println(topic, "<-", message.Message)
	return c.SendString("ok")
}

func Device(a *fiber.App) {
	// TODO: Checar se os dispositivos existem, mas a ideia é que tanto faz.

	a.Post("/dev/:id", func(c *fiber.Ctx) error {
		deviceId := c.Params("id")

		message := fmt.Sprintf("%v %s: %s", time.Now(), deviceId, string(c.Body()))

		// publica assim: torradeira: Temperatura=10C Umidade=20% Coisa=X

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
		messageChannel := make(chan string, 1)
		_, _ = pubsubber.Subscribe("/dev", messageChannel)
		return c.SendString(<-messageChannel)
	})
}

func WebClient(a *fiber.App) {
	a.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("static/view.html")
	})

}

func main() {
	app := fiber.New()

	Device(app)
	WebClient(app)

	app.Listen(":5000")
}
