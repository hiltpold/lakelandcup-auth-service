package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/service"
	"github.com/hiltpold/lakelandcup-auth-service/service/pb"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024
const testConfig = ".test.env"

var lis *bufconn.Listener

func init() {
	c, err := conf.LoadConfig(testConfig)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Failed to load config %s.", testConfig), err)
	}
	h := storage.Dial(&c.DB)

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

	loginReq := pb.LoginRequest{Password: "password", Email: "test@gmail.com"}

	resp, err := client.Login(ctx, &loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
	// Test for output here.
}

func TestRegister(t *testing.T) {
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

	resp, err := client.Register(ctx, &registerReq)

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
	// Test for output here.
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
