FROm golang
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build -o ./cmd/dotsd/main ./cmd/dotsd/main.go
EXPOSE 8080
CMD ["/app/cmd/dotsd/main"]
