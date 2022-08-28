package api

import (
	"context"
	"net/http"

	"github.com/hiltpold/lakelandcup-auth-service/api/pb"
	"github.com/hiltpold/lakelandcup-auth-service/models"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
)

type Server struct {
	C   storage.Connection
	Jwt utils.JwtWrapper
	// #https://github.com/grpc/grpc-go/issues/3794:
	pb.UnimplementedAuthServiceServer
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var user models.User

	if result := s.C.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error == nil {
		return &pb.RegisterResponse{
			Status: http.StatusConflict,
			Error:  "E-Mail already exists",
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

	s.C.DB.Create(&user)

	return &pb.RegisterResponse{
		Status: http.StatusCreated,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user models.User

	if result := s.C.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error != nil {
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

func (s *Server) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &pb.ValidateResponse{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}, nil
	}

	var user models.User

	if result := s.C.DB.Where(&models.User{Email: claims.Email}).First(&user); result.Error != nil {
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
