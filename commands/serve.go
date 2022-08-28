package commands

import (
	"log"
	"net"

	"github.com/hiltpold/lakelandcup-auth-service/api"
	"github.com/hiltpold/lakelandcup-auth-service/api/pb"
	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
	logger "github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
		SecretKey:       c.API.JWTSecretKey,
		Issuer:          "lakelandcup-auth-service",
		ExpirationHours: 24 * 365,
	}

	lis, err := net.Listen("tcp", c.API.Port)

	if err != nil {
		logger.Fatal("Failed to listing: ", zap.Error(err))
	}

	logger.Info("Lakelandcup Auth Service on Port: " + c.API.Port)

	s := api.Server{
		C:   h,
		Jwt: jwt,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
