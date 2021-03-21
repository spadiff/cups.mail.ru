FROM golang:1.16 AS build

WORKDIR /src
COPY . .
RUN go build -o goldrush .

FROM ubuntu
COPY --from=build /src/goldrush goldrush
CMD ["./goldrush"]