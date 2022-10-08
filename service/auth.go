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
			Error:  "Generate access token failed",
		}, nil
	}

	_, errSendMail := utils.SendGridMail(user.FirstName, user.Email, "Account Activation", "register", accessToken, os.Getenv("SENDGRID_KEY"))

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

func (s *Server) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	var user models.User

	if result := s.R.DB.Where(&models.User{Email: req.Email}).First(&user); result.Error != nil {
		return &pb.ForgotPasswordResponse{
			Status: http.StatusNotFound,
			Error:  "Email was never registered",
		}, nil
	}

	if !user.Confirmed {
		return &pb.ForgotPasswordResponse{
			Status: http.StatusForbidden,
			Error:  "Email was never activated",
		}, nil
	}

	forgotToken, errToken := s.Jwt.GenerateToken(utils.JwtData{Id: user.Id, Email: user.Email})

	if errToken != nil {
		defer logrus.Error(errToken.Error())
		return &pb.ForgotPasswordResponse{
			Status: http.StatusBadRequest,
			Error:  "Generate forgot access token failed",
		}, nil
	}

	_, errSendMail := utils.SendGridMail(user.FirstName, user.Email, "Reset Password", "fogot", forgotToken, os.Getenv("SENDGRID_KEY"))

	if errSendMail != nil {
		defer logrus.Error(errSendMail.Error())
		return &pb.ForgotPasswordResponse{
			Status: http.StatusBadRequest,
			Error:  "Sending email for retrieving password failed",
		}, nil
	}

	return &pb.ForgotPasswordResponse{
		Status: http.StatusOK,
	}, nil

}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	var user models.User

	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &pb.ResetPasswordResponse{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}, nil
	}

	if result := s.R.DB.Where(&models.User{Email: claims.Email}).First(&user); result.Error != nil {
		return &pb.ResetPasswordResponse{
			Status: http.StatusNotFound,
			Error:  "Email was never registered",
		}, nil
	}

	if req.Password != req.ConfirmPassword {
		return &pb.ResetPasswordResponse{
			Status: http.StatusForbidden,
			Error:  "Confirmation password does not match password",
		}, nil
	}

	user.Password, _ = utils.HashPassword(req.Password)

	if updateNewPassword := s.R.DB.Debug().Select("password", "update_at").Where("email = ?", claims.Email).Updates(user); updateNewPassword != nil {
		return &pb.ResetPasswordResponse{
			Status: http.StatusForbidden,
			Error:  "Error occured during password reset",
		}, nil

	}

	return &pb.ResetPasswordResponse{
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
