FROM golang:1.19-alpine3.16 AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod vendor -installsuffix cgo -o rollout-status github.com/socialgouv/rollout-status/cmd


FROM alpine:3.16

COPY --from=build /src/rollout-status /
ENTRYPOINT ["/rollout-status"]
