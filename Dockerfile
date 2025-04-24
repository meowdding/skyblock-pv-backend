FROM golang:1.24-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /skyblock-pv-backend
EXPOSE 8080
CMD ["/skyblock-pv-backend"]