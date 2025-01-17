# Cloud Run "Hello World" containers for REST / Cloud Event and gRPC

Google Cloud provides convenience container images for Cloud Run Services (`gcr.io/cloudrun/hello` / `us-docker.pkg.dev/cloudrun/container/hello`) and Cloud Run Jobs (`gcr.io/cloudrun/job` / `us-docker.pkg.dev/cloudrun/container/job`).
These prebuilt sample container images are used in Google Cloud Run tutorials.
These public container images, which require little configuration to pass health checks, are also useful when initializing Cloud Run Services and Jobs and bootstrapping Terraform projects.

~~However, the sample container image for Cloud Run Services only supports REST API and Cloud Event as the invocation types, and does not support gRPC invocation.~~

> [!NOTE]
> This is no longer true. Google Cloud now provides the gRPC hello world container image for Cloud Run at `us-docker.pkg.dev/cloudrun/container/hello`, and it is mentioned in [the official documentation on container health checks for Cloud Run](https://cloud.google.com/run/docs/configuring/healthchecks#terraform_4:~:text=image%0A%C2%A0%20%C2%A0%20%C2%A0%20image%20%3D%20%22-,us%2Ddocker.pkg.dev/cloudrun/container/hello,-%22%0A%0A%C2%A0%20%C2%A0%20%C2%A0%20liveness_probe%20%7B%0A%C2%A0%20%C2%A0%20%C2%A0%20%C2%A0%20failure_threshold%20%C2%A0%20%C2%A0).
> Skip the rest of this README and just grab the sample hello-world container image with gRPC health check installed from the Artifact Registry repository: [us-docker.pkg.dev/cloudrun/container/hello](https://console.cloud.google.com/artifacts/docker/cloudrun/us/container/hello).

This fork of the Cloud Run "Hello World" container source repository provides the source code of a sample gRPC server with the standard health check service enabled.
See the [/grpc](./grpc) directory for how to call this gRPC server.

A multi-arch Docker container image built from the gRPC server source code in the [/grpc](./grpc) directory is available on the Docker Hub: [`mmizutani/cloud-run-hello-grpc`](https://hub.docker.com/r/mmizutani/cloud-run-hello-grpc)

```bash
docker run -it --rm -p 50051:50051 -e PORT=50051 mmizutani/cloud-run-hello-grpc
```

Below is the original README in the [upstream repository](https://github.com/GoogleCloudPlatform/cloud-run-hello).

---

# Cloud Run "Hello" container

This repository contains the source code of a sample Go application that is
distributed as the public container image (`gcr.io/cloudrun/hello`) used in the
[Cloud Run quickstart](https://cloud.google.com/run/docs/quickstarts/) and as
the suggested container image in the Cloud Run UI on Cloud Console.

It also contains the source code of a placeholder public container
(`gcr.io/cloudrun/placeholder`)  used to create a placeholder revision when setting up
Continuous Deployment.

[![Run on Google Cloud](https://deploy.cloud.run/button.svg)](https://deploy.cloud.run)
