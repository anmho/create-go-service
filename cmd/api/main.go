package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/anmho/create-go-service/internal/api"
	"github.com/anmho/create-go-service/internal/database"
	"github.com/anmho/create-go-service/internal/notes"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	

	ctx := context.Background()

	dynamoClient, err := database.NewDynamoDB(ctx,
		database.WithEndpoint("http://localhost:8000"),
		database.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalln("failed to create dynamo client", err)
	}
	tableName := "NoteTable"
	results, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.Fatalln("failed to scan ddb table", err)
	}
	_ = results

	noteService := notes.NewService(dynamoClient, tableName)

	s := api.New(noteService)
	port := "8080"

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: s,
	}

	slog.Info("starting server", slog.String("port", port))
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}
