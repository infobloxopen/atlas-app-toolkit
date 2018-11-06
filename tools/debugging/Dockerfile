FROM golang:1.10.0 AS builder

LABEL stage=server-intermediate

WORKDIR /go/src/github.com/infobloxopen/protoc-gen-atlas-validate
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/usr/bin/protoc-gen-atlas-validate main.go

FROM infoblox/atlas-gentool:latest AS runner

COPY --from=builder /out/usr/bin/protoc-gen-atlas-validate /usr/bin/protoc-gen-atlas-validate

WORKDIR /go/src
