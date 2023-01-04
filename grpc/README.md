# Cloud Run "Hello World" container for gRPC

This is a sample gRPC server for deployment to Cloud Run Services.
The standard health check service is implemented, and the introspection API is enabled.

## How to launch

This sample gRPC server can be launched locally as follows:

```bash
go mod download
go run main.go
```


## How to call

This sample gRPC server can be called as follows:

- Health check service
  ```bash
  $ GRPC_HEALTH_PROBE_VERSION=v0.4.13 && \
      wget -qO grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
      chmod +x grpc_health_probe
  $ ./grpc_health_probe -addr 127.0.0.1:50051
  status: SERVING
  ```

  ```bash
  $ grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
  {
    "status": "SERVING"
  }
  ```

- Echo service
  ```bash
  $ grpcurl -plaintext -d '{"message":"gRPC"}' localhost:50051 grpc.examples.echo.Echo/UnaryEcho
  {
    "message": "gRPC"
  }
  ```

- Greeter service
  When run on the Cloud Run Services, the response would contain some information about your Google Cloud project.
  ```bash
  $ grpcurl -plaintext -d '{"name":"Tom"}' localhost:50051 helloworld.Greeter/SayHello
  {
    "message": "Hello Tom; service=<K_SERVICE>, revision=<K_REVISION, project=<PROJECT_ID>, region=<REGION>"
  }
  ```
