FROM "golang"

WORKDIR /app/renderer/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o entry

RUN chmod u+x entry

EXPOSE 80

CMD ["./entry"]
