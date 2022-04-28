FROM golang:1.18-alpine3.15 as build
WORKDIR /build
RUN apk add --no-cache --update git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o ./parse-plan

FROM alpine:3.15
COPY --from=build /build/parse-plan /app/parse-plan
ENTRYPOINT ["/app/parse-plan"]
