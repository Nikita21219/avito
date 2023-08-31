FROM golang:1.20.5-alpine3.17

WORKDIR /app

COPY app .

RUN mkdir -p csv_report_storage

RUN go build ./cmd/web

CMD ["./web"]
