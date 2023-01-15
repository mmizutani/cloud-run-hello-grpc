# Cloud Run "Hello World" container for gRPC

This is a sample gRPC server for deployment to Cloud Run Services.
The standard health check service is implemented, and the introspection API is enabled.

## How to launch

This sample gRPC server can be launched locally as follows:

```bash
go mod download
go run main.go
```


## How to call the gRPC server in the local environment

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

## Deploy to Cloud Run with locally built image

```bash
$ docker buildx create --name mybuilder --use --bootstrap
$ IMAGE_TAG="us-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY_NAME}/cloud-run-hello-grpc"
$ docker buildx build --push \
    --platform linux/amd64 \
    --tag $IMAGE_TAG .
$ gcloud run deploy $SERVICE --platform=managed --project=$PROJECT_ID --region=$REGION --image=$IMAGE_TAG
```

## How to call the gRPC server deployed on Cloud Run without the HTTP/2 cleartext (h2c) option (HTTP/2 requests are downgraded to HTTP/1.1 before reaching the service)

gRPC reflection requires HTTP/2 bidirectional steaming. So in order to enable gRPC reflection, it is required to connect the gRPC server via HTTP/2 over TCP (HTTP/2 cleartext or HTTP/2 without TLS aka h2c) and the gRPC server needs to support h2c connection.
Cloud Run, by default, downgrades HTTP/2 requests to HTTP/1.1, so it is necessary to disable this conversion by specifying `--use-http2` when deploying to Cloud Run.
However, the standard go library `net/http` does not expose easy configuration knobs for h2c, and correct implementation requires using the `golang.org/x/net/http2` package. Since this is beyond the scope of a basic gRPC server example, the implementation of gRPC reflection for Cloud Run deployment is dropped from this example.
Therefore, calling this gRPC server deployed on Cloud Run with the HTTP/2 cleartext (h2c) option disabled requires providing proto schema files as `grpccurl -proto <path_to_local_proto_file.proto> ...` even though the `-proto` option is not necessary when calling the server in the local environment.

### Deployment

```bash
gcloud run deploy $SERVICE --platform=managed --project=$PROJECT_ID --region=$REGION --source .
```



### Calling

```bash
$ FQDN=cloud-run-hello-grpc-genlookhtq-uc.a.run.app
$ PORT=443
$ grpcurl -proto protos/echo.proto -d '{"message":"gRPC"}' $FQDN:$PORT grpc.examples.echo.Echo/UnaryEcho
{
  "message": "gRPC"
}

$ grpcurl -proto protos/helloworld.proto -d '{"name":"Tom"}' $FQDN:$PORT helloworld.Greeter/SayHello
{
  "message": "Hello Tom; service=GreeterService, revision=cloud-run-hello-grpc-00007-mew, project=magic-modules-374220, regionus-central1"
}
```


## How to call the gRPC server deployed on Cloud Run with the HTTP/2 cleartext (h2c) option

Once the HTTP/2 cleartext (h2c) option is enabled on the Cloud Run service (`--use-http2`), we can use the server reflection as we do during local development and are not required to supply protobuf schema files in grpcurl commands.

### Deployment

```bash
gcloud run deploy $SERVICE --platform=managed --project=$PROJECT_ID --region=$REGION  --use-http2 --source .
```

### Calling

```bash
$ FQDN=$SERVICE-xxxxxxxxx-xx.a.run.app
$ PORT=443

$ grpcurl $FQDN:$PORT list
grpc.examples.echo.Echo
grpc.health.v1.Health
grpc.reflection.v1alpha.ServerReflection
helloworld.Greeter

$ grpcurl $FQDN:$PORT grpc.health.v1.Health/Check
{
  "status": "SERVING"
}

$ grpcurl -d '{"message":"gRPC"}' $FQDN:$PORT grpc.examples.echo.Echo/UnaryEcho
{
  "message": "gRPC"
}

$ grpcurl -d '{"name":"Tom"}' $FQDN:$PORT helloworld.Greeter/SayHello
{
  "message": "Hello Tom; service=GreeterService, revision=cloud-run-hello-grpc-00007-mew, project=magic-modules-374220, regionus-central1"
}
```

Instead of `grpcurl`, we can also use the [`buf curl`](https://docs.buf.build/curl/usage) command:

```bash
$ buf curl --protocol grpc -d '{"message":"gRPC"}' https://$SERVICE-xxxxxxxxx-xx.a.run.app/grpc.examples.echo.Echo/UnaryEcho
{"message":"gRPC"}

$ buf curl --protocol grpc -d '{"name":"Tom"}' https://$SERVICE-xxxxxxxxx-xx.a.run.app/helloworld.Greeter/SayHello
{"message":"Hello Tom; service=GreeterService, revision=$SERVICE-xxxxxxxxx-xx, project=$PROJECT_ID, region=$REGION"}
```
