package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	razorpay "github.com/razorpay/razorpay-go"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Payment struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email" bson:"email"`
	Phone     string `json:"phone" bson:"phone"`
	OrderID   string `json:"orderID" bson:"orderID"`
	PaymentID string `json:"paymentID" bson:"empaymentID"`
}

type OrderRequest struct {
	Amount int `json:"amount"`
}

var mongoClient *mongo.Client
var paymentCollection *mongo.Collection

func connectMongo() *mongo.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {

	app := fiber.New()

	// Allow CORS for all origins
	app.Use(cors.New())

	rzKey := goDotEnvVariable("RZKEY")
	rzPass := goDotEnvVariable("RZPASS")

	mongoClient = connectMongo()
	paymentCollection = mongoClient.Database("wym").Collection("payments")

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

	app.Get("/payment", func(c *fiber.Ctx) error {
		var payments []Payment
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := paymentCollection.Find(ctx, bson.M{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var payment Payment
			if err := cursor.Decode(&payment); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}
			payments = append(payments, payment)
		}

		return c.JSON(payments)
	})

	app.Post("/payment", func(c *fiber.Ctx) error {
		var payment Payment
		if err := c.BodyParser(&payment); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not insert payment"})
		}

		return c.Status(fiber.StatusCreated).JSON(payment)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local dev
	}

	log.Fatal(app.Listen(":" + port))
}
