package service

import (
	"auth_service/models"
	pb "auth_service/pb"
	"auth_service/repository"
	"context"
	"fmt"
	"log"

	petname "github.com/dustinkirkland/golang-petname"
)

type Service struct {
	pb.UnimplementedAuthServiceServer
	l    *log.Logger
	repo repository.IRepository
}

func NewService(l *log.Logger, repo repository.IRepository) Service {
	return Service{l: l, repo: repo}
}

func (s Service) Register(c context.Context, r *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	res := new(pb.RegisterResponse)
	user := models.User{}

	exists := true
	retryCount := 5

	for exists {
		// Check for the username in the cache, must faster
		user.Username = petname.Generate(2, "_")
		exists = s.repo.CheckUserNameExists(c, user.Username)

		if retryCount <= 0 {
			s.l.Printf("retrying username creating failed\n")
			res.Status = 500
			res.Error = "random username generation failed"
			return res, fmt.Errorf("random username generating failed.")
		}
		retryCount--
	}
	s.l.Printf("username selected %s\n", user.Username)
	user.Company = r.GetCompany()
	user.Password = r.GetPassword()

	err := s.repo.InsertUser(c, user)
	if err != nil {
		res.Status = 400
		res.Error = fmt.Sprintf("error creating user %s", err)
		return res, err
	}

	res.Status = 201
	res.Error = ""
	return res, nil

}
func (s Service) Login(c context.Context, r *pb.LoginRequest) (*pb.LoginResponse, error) {

	return nil, nil
}
func (s Service) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	return nil, nil
}
