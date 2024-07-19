FROM golang:1.22.5-alpine as builder
WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /main main.go

FROM alpine:3
COPY --from=builder main /bin/main
COPY employees.xlsx /app/employees.xlsx 
EXPOSE 8080
ENTRYPOINT ["bin/main"]