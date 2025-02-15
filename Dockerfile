FROM golang:1.23.4-alpine AS builder
ARG VERSION
ENV VERSION=${VERSION}
ARG APP_ENV
ENV APP_ENV=${APP_ENV}
RUN apk update && apk add --no-cache git
WORKDIR /gosm
COPY . /gosm/
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.version=${VERSION} -s -w" -o gosm cmd/restful/main.go

FROM scratch
COPY --from=builder /gosm/gosm .
COPY --from=builder /gosm/config.${APP_ENV}.yaml .
EXPOSE 8080
CMD [ "/gosm", "-env=${APP_ENV}" ]