FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY . .

ENV CGO_ENABLED=0

RUN go build -o dumper


FROM scratch

COPY --from=builder /src/dumper /dumper

ENTRYPOINT ["/dumper"]
