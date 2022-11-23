package commands

import (
	"fmt"
	"net"

	"github.com/hiltpold/lakelandcup-auth-service/conf"
	api "github.com/hiltpold/lakelandcup-auth-service/service"
	"github.com/hiltpold/lakelandcup-auth-service/service/pb"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		runWithConfig(cmd, serve)
	},
}

func serve(c *conf.Configuration) {
	h := storage.Dial(&c.DB)
	jwt := utils.JwtWrapper{
		TokenKey:            c.API.TokenSecretKey,
		TokenExpires:        c.API.TokenExpires,
		AccessTokenKey:      c.API.AccessTokenSecretKey,
		AccessTokenExpires:  c.API.AccessTokenExpires,
		RefreshTokenKey:     c.API.RefreshTokenSecretKey,
		RefreshTokenExpires: c.API.RefreshTokenExpires,
		Issuer:              "lakelandcup-auth-service",
		ExpirationHours:     24 * 365,
	}

	serviceUri := fmt.Sprintf(":%s", c.API.Port)

	lis, err := net.Listen("tcp", serviceUri)

	if err != nil {
		logrus.Fatal("Failed to listen on: ", err)
	}

	logrus.Info(fmt.Sprintf("Service [%s] from app [%s] is running on [%s]", c.API.Svc, c.API.App, serviceUri))

	s := api.Server{
		R:   h,
		Jwt: jwt,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		logrus.Fatalln("Failed to serve:", err)
	}
}
