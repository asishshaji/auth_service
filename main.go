package main

import (
	pb "auth_service/pb"
	"auth_service/repository"
	"auth_service/service"
	"auth_service/utils"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	logger := log.New(os.Stdout, "auth-service", log.LstdFlags)

	config := utils.LoadEnv(logger)
	db := config.InitDB()

	listener, err := net.Listen("tcp", config.APP_PORT)

	if err != nil {
		logger.Fatalf("error listen on port :%s, %s", config.APP_PORT, err)
	}

	server := grpc.NewServer()
	logger.Println("server started")
	repo := repository.NewPostgresRepo(logger, db)

	service := service.NewService(logger, repo)

	pb.RegisterAuthServiceServer(server, service)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("error starting grpc server: %s", err)
	}

	defer func() {
		db.Close()
		server.GracefulStop()
	}()
}
