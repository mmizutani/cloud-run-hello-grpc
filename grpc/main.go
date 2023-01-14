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
	"strconv"

	"google.golang.org/grpc"
	echopb "google.golang.org/grpc/examples/features/proto/echo"
	hwpb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

	port := 8080
	portStr := os.Getenv("PORT")
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("failed to parse PORT env var %s: %v", portStr, err)
		}
		port = portInt
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC health check service starts listening on port %d", port)

	s := grpc.NewServer()

	// Provides the gRPC service rpc grpc.health.v1.Health/Check
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s, healthcheck)

	// Provides the gRPC service rpc grpc.examples.echo.Echo/UnaryEcho
	echopb.RegisterEchoServer(s, &echoServer{})

	// Provides the gRPC service rpc helloworld.Greeter/SayHello
	greeter := &greeterServer{}
	data.Service = "GreeterService"
	greeter.data = data
	hwpb.RegisterGreeterServer(s, greeter)

	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
