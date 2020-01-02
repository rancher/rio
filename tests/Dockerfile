FROM golang:1.13-alpine

# Install all needed tools
RUN apk update && \
  apk upgrade --update-cache --available && \
  apk add build-base curl git jq openssh bash docker

# Install the executables for kubectl, rio, and hey
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
  chmod +x ./kubectl && \
  mv ./kubectl /usr/local/bin && \
  go get -u github.com/rakyll/hey

# Set working directory to rio application directory and install all go dependencies
WORKDIR /usr/local/projects/rio
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source code into working directory and make sure entrypoint is executable
COPY . .
RUN chmod +x ./tests/scripts/*.sh && chmod +x ./scripts/build

# Set environment variables for tests to run properly and to ensure the correct rio version is being tested
ENV CLUSTER local
ENV KUBECONFIG /usr/local/projects/rio/.kube/config
ENV test integration

# Install Rio if needed and run tests
## NOTE: Need to pass in environment variables:
 # RIO_VERSION  (optional - pass 'master' if wanting to build from source)
 # REPO         (required if RIO_VERSION=master. Specify docker username for rio_controller)
 # TAG          (required if RIO_VERSION=master. Specify rio_controller tag. Ends up like: ${REPO}/rio-controller:${TAG})
 # CLUSTER      (optional - if wanting to build and use a new cluster. Valid options are 'k3s', 'rke', and 'gke')
 # TOKEN        (required if CLUSTER is given as k3s or rke -- DigitalOcean API Token)
 # WORKERS      (optional - if passing in the cluster option, you can specify how many additional worker nodes to add)
ENTRYPOINT [ "./tests/scripts/entrypoint.sh" ]
CMD ["./tests/scripts/test.sh"]
