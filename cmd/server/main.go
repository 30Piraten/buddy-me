package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	checkpointv1 "github.com/30Piraten/buddy-backend/gen/go/proto/checkpoints/v1"
	roadmapv1 "github.com/30Piraten/buddy-backend/gen/go/proto/roadmaps/v1"
	usersv1 "github.com/30Piraten/buddy-backend/gen/go/proto/users/v1"
	checkpointgen "github.com/30Piraten/buddy-backend/internal/db/checkpoints/checkpoint_generated"
	roadmapgen "github.com/30Piraten/buddy-backend/internal/db/roadmaps/roadmap_generated"
	usergen "github.com/30Piraten/buddy-backend/internal/db/users/user_generated"
	"github.com/30Piraten/buddy-backend/internal/handlers/checkpoints"
	"github.com/30Piraten/buddy-backend/internal/handlers/roadmap"
	"github.com/30Piraten/buddy-backend/internal/handlers/users"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/30Piraten/buddy-backend/internal/logging"
	"github.com/rs/zerolog/log"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		if fallback != "" {
			log.Info().Str("WARNING: %s environment variable is required", key)
			return fallback
		}
		log.Info().Str("ERROR: %s environment variable is required", key)
		return ""
	}
	return value
}

func main() {

	logging.Init()
	log.Info().Msg("Logger initialised")

	err := godotenv.Load()
	if err != nil {
		log.Info().Str("Warning", "Error loading .env file: "+err.Error())
		log.Info().Msg("Continuing without .env file - will use environment variables or fallback")
	} else {
		log.Info().Msg("Successfully loaded .env file")
	}

	// Using environment variable with fallback
	dbConn := getEnv("POSTGRES_DSN", "postgres://buddy:secret@localhost:5432/buddy?sslmode=disable")

	log.Info().Str("Connecting to buddy-backend database %s:", dbConn)

	// Create a context with timeout for database connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure connection pool
	config, err := pgxpool.ParseConfig(dbConn)
	if err != nil {
		log.Info().Str("Error", "Failed to parse database connection string: "+err.Error())
	}

	// Set connection pool params
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Connect to the database
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Info().Str("Error", "Failed to connect to database: "+err.Error())
	}
	defer db.Close()

	// Ping the database to verify if connection works
	if err := db.Ping(ctx); err != nil {
		log.Info().Str("Error", "Failed to ping databse: "+err.Error())
	}
	log.Info().Msg("Successfully hitchhiked to the database")

	//

	// Users
	userQueries := usergen.New(db)
	userHandler := users.NewUserHandler(userQueries)

	// Roadmaps
	roadmapQueries := roadmapgen.New(db)
	roadmapHandler := roadmap.NewRoadmapHandler(roadmapQueries)

	// Checkpoint
	checkpointQueries := checkpointgen.New(db)
	checkpointHandler := checkpoints.NewCheckpointHandler(checkpointQueries)

	//

	port := getEnv("PORT", "9090")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Info().Str("Error", "Failed to connect to port: "+err.Error())
	}

	//

	// Users
	server := grpc.NewServer()
	usersv1.RegisterUserServiceServer(server, userHandler)

	// Roadmap
	roadmapv1.RegisterRoadmapServiceServer(server, roadmapHandler)

	// Checkpoint
	checkpointv1.RegisterCheckpointServiceServer(server, checkpointHandler)

	//

	// Register reflection service on gRPC server
	reflection.Register(server)

	log.Info().Msg("gRPC server trekking on :9090")
	if err := server.Serve(listener); err != nil {
		log.Info().Str("Error", "Failed to connect to server: "+err.Error())
	}

}
