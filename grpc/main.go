package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"net/http"
	"os"
	"regexp"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	echopb "google.golang.org/grpc/examples/features/proto/echo"
	hwpb "google.golang.org/grpc/examples/helloworld/helloworld"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	// "google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Data struct {
	Service  string
	Revision string
	Project  string
	Region   string
}

type echoServer struct {
	echopb.UnimplementedEchoServer
}

func (e *echoServer) UnaryEcho(ctx context.Context, req *echopb.EchoRequest) (*echopb.EchoResponse, error) {
	return &echopb.EchoResponse{Message: req.Message}, nil
}

type greeterServer struct {
	hwpb.UnimplementedGreeterServer
	data Data
}

func (s *greeterServer) SayHello(ctx context.Context, req *hwpb.HelloRequest) (*hwpb.HelloReply, error) {
	return &hwpb.HelloReply{
		Message: fmt.Sprintf(
			"Hello %s; service=%s, revision=%s, project=%s, region%s",
			req.Name,
			s.data.Service,
			s.data.Revision,
			s.data.Project,
			s.data.Region,
		),
	}, nil
}

func main() {
	// Get project ID from metadata server
	project := ""
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			responseBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			project = string(responseBody)
		}
	}

	// Get region from metadata server
	region := ""
	req, _ = http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/region", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	res, err = client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			responseBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			region = regexp.MustCompile(`projects/[^/]*/regions/`).ReplaceAllString(string(responseBody), "")
		}
	}
	if region == "" {
		// Fallback: get "zone" from metadata server (running on VM e.g. Cloud Run for Anthos)
		req, _ = http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
		req.Header.Set("Metadata-Flavor", "Google")
		res, err = client.Do(req)
		if err == nil {
			defer res.Body.Close()
			if res.StatusCode == 200 {
				responseBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Fatal(err)
				}
				region = regexp.MustCompile(`projects/[^/]*/zones/`).ReplaceAllString(string(responseBody), "")
			}
		}
	}

	service := os.Getenv("K_SERVICE")
	revision := os.Getenv("K_REVISION")

	data := Data{
		Service:  service,
		Revision: revision,
		Project:  project,
		Region:   region,
	}

	zapOpt := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	grpc_zap.ReplaceGrpcLogger(logger)

	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(logger, zapOpt...),
		),
	)

	// Provides the gRPC service rpc grpc.health.v1.Health/Check
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(server, healthcheck)

	// Provides the gRPC service rpc grpc.examples.echo.Echo/UnaryEcho
	echopb.RegisterEchoServer(server, &echoServer{})

	// Provides the gRPC service rpc helloworld.Greeter/SayHello
	greeter := &greeterServer{}
	data.Service = "GreeterService"
	greeter.data = data
	hwpb.RegisterGreeterServer(server, greeter)

	reflection.Register(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC services start listening on port %s", port)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
