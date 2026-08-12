FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/faker .

FROM scratch

COPY --from=build /out/faker /faker

USER 65532:65532

EXPOSE 27016/udp

ENTRYPOINT ["/faker"]
CMD ["-address=game-server:27016", "-port=:27016"]
