package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed" bson:"completed"`
	Body      string             `json:"body" bson:"body"`
}

var collection *mongo.Collection

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	MONGODB_URI := os.Getenv("MONGO_DB_URI")
	PORT := os.Getenv("PORT")

	clientOptions := options.Client().ApplyURI((MONGODB_URI))

	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	fmt.Println("Connected to mongodb...")

	collection = client.Database("gotasks").Collection("todos")

	app := fiber.New()

	api := app.Group("/api")
	apiV1 := api.Group("/v1")

	apiV1.Get("/todos", getTodos)
	apiV1.Post("/todos/create", createTodo)
	apiV1.Put("/todos/update/:id", updateTodo)
	apiV1.Delete("/todos/delete/:id", deleteTodo)

	log.Fatal(app.Listen(":" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	var todos = []Todo{}

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}

		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	// how you create a pointer
	todo := new(Todo)

	err := c.BodyParser(todo)
	if err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Todo body cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	todoId := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(todoId)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid todo id"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid todo id"})
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true})
}
