package service

import (
	"context"
	"net/http"
	"os"

	"github.com/hiltpold/lakelandcup-auth-service/models"
	"github.com/hiltpold/lakelandcup-auth-service/service/pb"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/sirupsen/logrus"
)

type Server struct {
	R   storage.Repository
	Jwt utils.JwtWrapper
	// #https://github.com/grpc/grpc-go/issues/3794:
	pb.UnimplementedAuthServiceServer
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var user models.User

	if findUser := s.R.DB.Where(&models.User{Email: req.Email}).First(&user); findUser.Error == nil {
		return &pb.RegisterResponse{
			Status: http.StatusConflict,
			Error:  "Email already exists",
		}, nil
	}

	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	password, err := utils.HashPassword(req.Password)

	if err != nil {
		return nil, err
	}
	user.Password = password

	if createUser := s.R.DB.Create(&user); createUser.Error != nil {
		return &pb.RegisterResponse{
			Status: http.StatusForbidden,
			Error:  "Register new account failed",
		}, nil
	}

	accessToken, errToken := s.Jwt.GenerateToken(utils.JwtData{Id: user.Id, Email: user.Email})

	if errToken != nil {
		defer logrus.Error(errToken.Error())
		return &pb.RegisterResponse{
			Status: http.StatusBadRequest,
			Error:  "Generate accessToken failed",
		}, nil
	}

	_, errSendMail := utils.SendGridMail(user.FirstName, user.Email, "Activation Account", "register", accessToken, os.Getenv("SENDGRID_KEY"))

	if errSendMail != nil {
		defer logrus.Error(errSendMail.Error())
		return &pb.RegisterResponse{
			Status: http.StatusBadRequest,
			Error:  "Sending email activation failed",
		}, nil
	}

	return &pb.RegisterResponse{
		Status: http.StatusCreated,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user models.User

	if result := s.R.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error != nil {
		return &pb.LoginResponse{
			Status: http.StatusNotFound,
			Error:  "Incorrect email or password",
		}, nil
	}

	match := utils.CheckPasswordHash(req.Password, user.Password)

	if !match {
		return &pb.LoginResponse{
			Status: http.StatusNotFound,
			Error:  "Incorrect email or password",
		}, nil
	}

	token, _ := s.Jwt.GenerateToken(utils.JwtData{Id: user.Id, Email: user.Email})

	return &pb.LoginResponse{
		Status: http.StatusOK,
		Token:  token,
	}, nil
}

func (s *Server) Activate(ctx context.Context, req *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	var user models.User

	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &pb.ActivateResponse{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}, nil
	}

	if result := s.R.DB.Where(&models.User{Email: claims.Email}).First(&user); result.Error != nil {
		return &pb.ActivateResponse{
			Status: http.StatusNotFound,
			Error:  "Token does not belong to a user",
		}, nil
	}

	if update := s.R.DB.Model(&user).Update("Confirmed", true); update.Error != nil {
		return &pb.ActivateResponse{
			Status: http.StatusNotFound,
			Error:  "User could not be updated",
		}, nil
	}

	return &pb.ActivateResponse{
		Status: http.StatusOK,
	}, nil
}

func (s *Server) ResendActivationToken(ctx context.Context, req *pb.ResendActivationTokenRequest) (*pb.ResendActivationTokenResponse, error) {
	var user models.User

	if result := s.R.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error != nil {
		return &pb.ResendActivationTokenResponse{
			Status: http.StatusNotFound,
			Error:  "Email was never registered",
		}, nil
	}

	accessToken, errToken := s.Jwt.GenerateToken(utils.JwtData{Id: user.Id, Email: user.Email})

	if errToken != nil {
		defer logrus.Error(errToken.Error())
		return &pb.ResendActivationTokenResponse{
			Status: http.StatusBadRequest,
			Error:  "Generate accessToken failed",
		}, nil
	}

	_, errSendMail := utils.SendGridMail(user.FirstName, user.Email, "Activation Account", "register", accessToken, os.Getenv("SENDGRID_KEY"))

	if errSendMail != nil {
		defer logrus.Error(errSendMail.Error())
		return &pb.ResendActivationTokenResponse{
			Status: http.StatusBadRequest,
			Error:  "Sending email activation failed",
		}, nil
	}

	return &pb.ResendActivationTokenResponse{
		Status: http.StatusOK,
	}, nil
}

func (s *Server) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &pb.ValidateResponse{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}, nil
	}

	var user models.User

	if result := s.R.DB.Where(&models.User{Email: claims.Email}).First(&user); result.Error != nil {
		return &pb.ValidateResponse{
			Status: http.StatusNotFound,
			Error:  "User not found",
		}, nil
	}

	return &pb.ValidateResponse{
		Status: http.StatusOK,
		UserId: user.Id.String(),
	}, nil
}
