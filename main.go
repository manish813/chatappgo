package main

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	connections = make(map[string]*websocket.Conn)
	mutex       sync.Mutex
)

func main() {

	app := fiber.New()

	app.Use(cors.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {

		id := c.Params("id")
		mutex.Lock()
		connections[id] = c
		mutex.Unlock()

		log.Printf("Connection established for Id : %s \n", id)

		defer func() {
			mutex.Lock()
			delete(connections, id)
			mutex.Unlock()
			log.Printf("Connections closed for Id : %s \n", id)
		}()

		var (
			mt  int
			msg []byte
			err error
		)

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("Recieved message from : %s %s \n", id, msg)

			broadcastMessage(id, msg, mt)

		}

	}))

	log.Fatal(app.Listen(":3000"))
}

func broadcastMessage(senderId string, msg []byte, mt int) {
	mutex.Lock()
	defer mutex.Unlock()

	for id, conn := range connections {
		if id != senderId {
			if err := conn.WriteMessage(mt, msg); err != nil {
				log.Printf("Failed to send message %s: %v \n", id, err)
			}
		}
	}
}
