FROm alpine:latest
RUN mkdir -p /app/tmp
WORKDIR /app
COPY ./api .
EXPOSE 5432
