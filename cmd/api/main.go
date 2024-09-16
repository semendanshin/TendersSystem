package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
	"strings"
	"tenderSystem/internal/infrastructure/repositories/bid"
	"tenderSystem/internal/infrastructure/repositories/bid/decision"
	"tenderSystem/internal/infrastructure/repositories/bid/feedback"
	"tenderSystem/internal/infrastructure/repositories/employee"
	"tenderSystem/internal/infrastructure/repositories/tender"
	"tenderSystem/internal/infrastructure/server"
	"tenderSystem/internal/usecase"

	"github.com/joho/godotenv"
)

func inner() error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUsername := os.Getenv("POSTGRES_USERNAME")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDatabase := os.Getenv("POSTGRES_DATABASE")

	postgresURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", postgresHost, postgresPort, postgresUsername, postgresPassword, postgresDatabase)

	// Split the server address into host and port
	var host, port string
	addressParts := strings.SplitN(serverAddress, ":", 2)
	host, port = addressParts[0], addressParts[1]

	// Connect to the database
	pgxConn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return err
	}
	defer pgxConn.Close(context.Background())

	// Ping the database
	err = pgxConn.Ping(context.Background())
	if err != nil {
		return err
	}

	// Init repositories
	tenderRepo := tender.NewPGXRepository(pgxConn)
	bidRepo := bid.NewPGXRepository(pgxConn)
	bidFeedbackRepo := feedback.NewPGXRepository(pgxConn)
	bidDecisionRepo := decision.NewPGXRepository(pgxConn)

	employeeRepo := employee.NewPGXRepository(pgxConn)

	// Init use cases
	tenderUseCase := usecase.NewTenderUseCase(tenderRepo, employeeRepo)
	bidUseCase := usecase.NewBidUseCase(employeeRepo, tenderRepo, bidRepo, bidFeedbackRepo, bidDecisionRepo)

	// Init server
	srv := server.NewServer(tenderUseCase, bidUseCase, host, port)

	return srv.Start()
}

func main() {
	err := inner()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
