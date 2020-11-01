# build container image
FROM golang:alpine as builder

ARG SERVICE_NAME="from_cmdline"
ARG SERVICE_PORT=0
ARG VERSION="from_cmdline"
ARG BUILD_INFO="from_cmdline"

#RUN mkdir /build
WORKDIR /build
RUN apk update && apk add --no-cache git
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

RUN GO111MODULE=on GOOS=linux CGO_ENABLED=0 \
go build \
-ldflags "-X main.version=${VERSION} -X 'main.buildInfo=${BUILD_INFO}'" \
-o server ./cmd/${SERVICE_NAME}

# release container image
FROM alpine:latest 

ARG SERVICE_NAME="from_cmdline"
ARG SERVICE_PORT=8000

WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/server .

# expose listening port
EXPOSE ${SERVICE_PORT}
ENV PORT=${SERVICE_PORT}

# run server
CMD [ "./server" ]