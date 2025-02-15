FROM golang:1.23.4-alpine AS builder
ARG VERSION
ENV VERSION=${VERSION}
RUN apk update && apk add --no-cache git
WORKDIR /gosm
COPY . /gosm/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.version=${VERSION} -s -w" -o gosm cmd/restful/main.go

FROM scratch
COPY --from=builder /gosm/gosm .
COPY --from=builder /gosm/config.prod.yaml .
EXPOSE 8080
CMD [ "/gosm", "-env=prod" ]