# Copyright 2026 Zyvor AI Labs
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/chimera ./cmd/chimera

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/chimera /usr/local/bin/chimera
EXPOSE 8989
ENTRYPOINT ["/usr/local/bin/chimera","serve"]
CMD ["-listen","0.0.0.0:8989"]
