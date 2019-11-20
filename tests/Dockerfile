FROM golang:1.13-alpine

# Install all needed tools
RUN apk update && \
  apk upgrade --update-cache --available && \
  apk add build-base curl git

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
RUN chmod +x ./tests/entrypoint.sh

# Set environment variables for tests to run properly and to ensure the correct rio version is being tested
ENV KUBECONFIG /usr/local/projects/rio/.kube/config
ENV test integration

# Install Rio if needed and run tests
ENTRYPOINT [ "./tests/entrypoint.sh" ]
CMD go test -v ./tests/${test}/... -${test}-tests
