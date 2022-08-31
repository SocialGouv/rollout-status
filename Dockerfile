FROM golang:1.12-alpine3.10 AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -mod vendor -installsuffix cgo -o rollout-status github.com/socialgouv/rollout-status/cmd


FROM alpine:3.10

COPY --from=build /src/rollout-status /
ENTRYPOINT ["/rollout-status"]
