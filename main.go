package main

import (
	"log"

	"sbm.gg/verified-users/server"
	"sbm.gg/verified-users/mongo"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Server...")
	loadEnv()
	mongo.Init()
	server.Init()
}

func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Env loaded successfully")
}