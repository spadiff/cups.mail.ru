FROM golang:1.16 AS build

WORKDIR /src
COPY . .
ENV GOMAXPROCS=4
RUN go build -o goldrush .

FROM ubuntu
COPY --from=build /src/goldrush goldrush
ENV GOMAXPROCS=4
CMD ["./goldrush"]