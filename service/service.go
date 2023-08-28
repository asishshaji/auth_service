package service

import (
	"auth_service/models"
	pb "auth_service/pb"
	"auth_service/repository"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"

	petname "github.com/dustinkirkland/golang-petname"
	"golang.org/x/crypto/bcrypt"
)

type serviceOpts struct {
	JWT_Secret string
	Salt       string
}

type Service struct {
	l     *log.Logger
	repo  repository.IRepository
	redis *redis.Client
	serviceOpts
	pb.UnimplementedAuthServiceServer
}

func NewService(l *log.Logger, repo repository.IRepository, redis *redis.Client) Service {
	return Service{l: l, repo: repo, redis: redis}
}

func (s Service) Register(c context.Context, r *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	res := new(pb.RegisterResponse)
	user := models.User{}

	retryCount := 5
	exists := true

	var err error

	for exists {
		// Check for the username in the cache, must faster
		user.Username = petname.Generate(2, "_")
		exists, err = s.repo.CheckUserNameExists(c, user.Username)

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.GetPassword()), 5)
	if err != nil {
		res.Status = 500
		res.Error = fmt.Sprintf("error generating password hash: %s", err)
		return res, err
	}
	user.Password = string(hashedPassword)

	err = s.repo.InsertUser(c, user)
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
	res := new(pb.LoginResponse)
	hashedPassword, err := s.repo.GetUserPassword(c, r.GetUsername())
	if err != nil {
		res.Status = 404
		res.Error = fmt.Sprintf("error getting user : %s", err)
		return res, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(r.GetPassword()))
	if err != nil {
		res.Status = 403
		res.Error = fmt.Sprintf("invalid password : %s", err)
		return res, err
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
	claims["user"] = r.GetUsername()
	claims["authorized"] = true

	tokenStr, err := token.SignedString([]byte(s.JWT_Secret))
	if err != nil {
		fmt.Println(err)
		res.Status = 500
		res.Error = fmt.Sprintf("error generating access token : %s", err)
		return res, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)

	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshClaims["exp"] = time.Now().Add(10 * time.Hour * 24)
	refreshClaims["user"] = r.GetUsername()
	refreshClaims["authorized"] = true

	refreshTokenStr, err := refreshToken.SignedString([]byte(s.JWT_Secret))
	if err != nil {
		fmt.Println(err)
		res.Status = 500
		res.Error = fmt.Sprintf("error generating refresh token : %s", err)
		return res, err
	}

	res.AccessToken = tokenStr
	res.RefreshToken = refreshTokenStr
	res.Status = 200
	res.Error = ""

	// TODO persistence for session
	// https://redis.com/blog/json-web-tokens-jwt-are-dangerous-for-user-sessions/
	s.redis.SAdd(context.Background(), "users:auth_issued", r.GetUsername())

	return res, nil
}

func (s Service) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	res := new(pb.ValidateResponse)
	tokenStr := r.GetToken()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.JWT_Secret), nil
	})

	if err != nil {
		s.l.Println(err)
		res.Error = fmt.Sprintf("token parsing failed : %s", err)
		res.Status = 403
		return res, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.l.Println(err)
		res.Error = fmt.Sprintf("claims parsing failed : %s", err)
		res.Status = 403
		return res, err
	}
	username, ok := claims["user"].(string)
	if !ok {
		s.l.Println(err)
		res.Error = fmt.Sprintf("username parsing failed : %s", err)
		res.Status = 403
		return res, err
	}

	is := s.redis.SIsMember(c, "users:auth_issued", username)
	if !is.Val() {
		res.Error = "user not authorized"
		res.Status = 403
		return res, fmt.Errorf("user not authorized")
	}

	res.Status = 200

	return nil, nil
}
