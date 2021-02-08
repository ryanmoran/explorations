FROM golang:1.15
WORKDIR /app
ADD . .
RUN go install ./
ENTRYPOINT go-app
