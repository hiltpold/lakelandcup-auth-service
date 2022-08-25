package commands

import (
	"fmt"
	"log"
	"net"

	"github.com/hiltpold/lakelandcup-auth-service/api"
	"github.com/hiltpold/lakelandcup-auth-service/api/pb"
	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/storage"
	"github.com/hiltpold/lakelandcup-auth-service/utils"
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
	h := storage.Dial(c.URI)

	jwt := utils.JwtWrapper{
		SecretKey:       c.JWTSecretKey,
		Issuer:          "lakelandcup-auth-service",
		ExpirationHours: 24 * 365,
	}

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Auth Svc on", c.Port)

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
