package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/models"
	"github.com/hiltpold/lakelandcup-auth-service/service"
	"github.com/hiltpold/lakelandcup-auth-service/service/pb"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"
)

const bufSize = 1024 * 1024
const testConfig = ".test.env"

var lis *bufconn.Listener
var db *gorm.DB

func init() {
	c, err := conf.LoadConfig(testConfig)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Failed to load config %s.", testConfig), err)
	}
	h := storage.Dial(&c.DB)
	db = h.DB
	jwt := utils.JwtWrapper{
		SecretKey:       c.API.JWTSecretKey,
		Issuer:          "lakelandcup-auth-service",
		ExpirationHours: 24 * 365,
	}

	lis = bufconn.Listen(bufSize)
	s := service.Server{
		R:   h,
		Jwt: jwt,
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &s)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewAuthServiceClient(conn)

	loginReq := pb.LoginRequest{Email: "max.muster@gmail.com", Password: "password"}

	loginResp, err := client.Login(ctx, &loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	log.Printf("Response: %+v", loginResp)

	// Test for response here.
	assert.Equal(t, loginResp.Status, int64(404))
	assert.Equal(t, loginResp.Error, "Incorrect email or password")
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewAuthServiceClient(conn)

	registerReq1 := pb.RegisterRequest{
		FirstName: "Max",
		LastName:  "Muster",
		Email:     "max.muster@gmail.com",
		Password:  "password",
	}

	registerResp1, err1 := client.Register(ctx, &registerReq1)

	if err1 != nil {
		t.Fatalf("Login failed: %v", err1)
	}

	log.Printf("Response: %+v", registerResp1)
	// Test for response here.
	assert.Equal(t, registerResp1.Status, int64(201))

	// Clean Up
	user := models.User{}
	if result := db.Where(&models.User{Email: registerReq1.Email}).First(&user); result.Error != nil {
		logrus.Fatal("Could not query Database", result.Error)
	}
	db.Delete(&user)

}

func TestRegisterAndLogin(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewAuthServiceClient(conn)

	registerReq := pb.RegisterRequest{
		FirstName: "Max",
		LastName:  "Muster",
		Email:     "max.muster@gmail.com",
		Password:  "password",
	}

	registerResp, err := client.Register(ctx, &registerReq)
	log.Printf("Response: %+v", registerResp)

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	loginReq := pb.LoginRequest{Password: "password", Email: "test@gmail.com"}

	loginResp, err2 := client.Login(ctx, &loginReq)
	log.Printf("Response: %+v", loginResp)

	if err2 != nil {
		t.Fatalf("Login failed: %v", err2)
	}
	// Test for output here.
}
