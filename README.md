# Fractionable US Stocks Price Checker (Go)

`Alpaca`에서 미국 주식 중 분할매매 가능(`fractionable=true`) 종목 목록을 가져오고,
`Yahoo Finance`에서 해당 종목의 현재가를 조회하는 CLI 프로그램입니다.

## 요구 사항

- Go 1.22+
- Alpaca API Key/Secret

## 환경 변수

```bash
export APCA_API_KEY_ID="YOUR_KEY"
export APCA_API_SECRET_KEY="YOUR_SECRET"
```

## 실행

```bash
go run .
```

옵션:

- `--limit 50`: 출력 종목 개수 제한 (기본 30)
- `--symbols AAPL,MSFT,TSLA`: 특정 심볼만 필터링
- `--format json`: JSON 출력 (`table` 또는 `json`)
- `--alpaca-url https://paper-api.alpaca.markets`: Alpaca API URL 변경

예시:

```bash
go run . --limit 10 --symbols AAPL,MSFT,NVDA
```

## 테스트

```bash
go test ./...
```

## 참고

- 분할매매 가능 여부는 브로커 정책에 따라 달라질 수 있습니다.
- 이 프로그램은 Alpaca의 `fractionable` 정보를 기준으로 동작합니다.
