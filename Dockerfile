FROm golang
RUN mkdir /app
ADD ./ /app
WORKDIR /app
RUN go build -o ./cmd/dotsd/main ./cmd/dotsd/main.go
EXPOSE 8080
CMD ["/app/cmd/dotsd/main"]
