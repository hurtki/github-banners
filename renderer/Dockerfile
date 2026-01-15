FROM "golang"

WORKDIR /app/crud/

COPY . .

RUN go build -o entry

RUN chmod u+x entry

CMD ["./entry"]
