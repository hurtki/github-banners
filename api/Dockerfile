FROM "golang" AS build

WORKDIR /app/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o entry

RUN chmod u+x entry

EXPOSE 80

FROM alpine:latest

COPY --from=build /app/entry .

CMD ["./entry"]
