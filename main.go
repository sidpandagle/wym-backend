package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	razorpay "github.com/razorpay/razorpay-go"
)

func main() {

	app := fiber.New()
	rzKey := goDotEnvVariable("RZKEY")
	rzPass := goDotEnvVariable("RZPASS")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Post("/order", func(c *fiber.Ctx) error {
		payload := new(OrderRequest)
		if err := c.BodyParser(payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		client := razorpay.NewClient(rzKey, rzPass)

		data := map[string]interface{}{
			"amount":   payload.Amount * 100,
			"currency": "INR",
			"receipt":  "some_receipt_id",
		}
		body, err := client.Order.Create(data, nil)
		if err != nil {
			panic(err)
		}
		return c.JSON(body)
	})

	log.Fatal(app.Listen(":3000"))
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

type OrderRequest struct {
	Amount int `json:"amount"`
}
