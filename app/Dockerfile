FROM golang:1.14 as builder

RUN mkdir /app

COPY . /app

WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .


FROM scratch

ENV PORT=8080

COPY --from=builder /app/server /

EXPOSE 8080
CMD ["/server"]
