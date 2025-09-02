# ---------- build stage ----------
FROM golang:1.23 AS build
WORKDIR /app

# 依存解決をキャッシュさせる
COPY go.mod ./
RUN go mod download

# 残りのソース
COPY . .

# 本体バイナリ（ここでは "work-tracker" を想定）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/work-tracker ./main/work-tracker

# ---------- runtime stage ----------
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/work-tracker /usr/local/bin/work-tracker
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/work-tracker"]
