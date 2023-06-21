FROm alpine:latest
RUN mkdir -p /app/tmp
WORKDIR /app
COPY . .
EXPOSE 5432
