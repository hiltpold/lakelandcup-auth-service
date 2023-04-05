# Stage 1: compile the program
FROM golang:latest as build
WORKDIR /app
COPY go.* /app/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o lakelandcup-auth-service main.go

# Stage 2: build the image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /app/
COPY --from=build /app/lakelandcup-auth-service .
COPY --from=build /app/.prod.env /app/.prod.env
EXPOSE 50010
CMD ["./lakelandcup-auth-service","-c",".prod.env"]  