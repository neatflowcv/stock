# USEquityMarket

## 정의

`USEquityMarket`은 미국 주식이 거래되고 상장되는 시장 구조 전체를 뜻한다.

한국의 `KOSPI/KOSDAQ`처럼 2개 시장으로 단순 구분하기보다,
`거래소(Exchange)`, `상장 시장 Tier`, `장외시장(OTC)`,
`거래 체결 Venue`를 분리해서 이해해야 한다.

## 핵심 용어

### Exchange

- 주식이 공식적으로 상장(listing)되는 거래소
- 대표: `NYSE`, `NASDAQ`, `NYSE American`

### Listed Exchange

- 특정 종목의 주 상장 거래소
- 예: `AAPL`의 주 상장 거래소는 `NASDAQ`

### Primary Listing

- 기업이 공식적으로 기준 삼는 상장
- 다중 상장(multiple listing)이 있어도 통상 1개를 주 상장으로 본다

### Market Tier (NASDAQ)

- NASDAQ 내부 상장 구분
- `NASDAQ Global Select Market`
- `NASDAQ Global Market`
- `NASDAQ Capital Market`

### Market Tier (NYSE 계열)

- `NYSE`, `NYSE American` 등으로 상장 요건과 성격이 나뉜다
- 한국의 코스피/코스닥처럼 단일 브랜드 2분할이 아니다

### Listing Standards

- 상장 유지에 필요한 기준
- 시가총액, 주가, 유동주식수, 재무요건, 공시요건 등을 포함

### Delisting

- 상장 폐지
- 자진 상폐(voluntary) 또는 요건 미달(non-compliance)로 발생

### OTC (Over-The-Counter)

- 거래소 상장 없이 브로커 네트워크 중심으로 거래되는 시장
- 일반적으로 `NYSE/NASDAQ` 대비 정보 접근성과 유동성이 낮을 수 있다

### Ticker Symbol

- 종목 코드
- 미국에서는 거래소/데이터 벤더에 따라 표기 변형이 있을 수 있다
- 예: 클래스주 점 표기(`BRK.B`, `BRK-B`) 차이

### Share Class

- 같은 기업의 서로 다른 의결권/권리 구조 주식 클래스
- 예: `GOOG` vs `GOOGL`

### ADR (American Depositary Receipt)

- 해외 기업 주식을 미국 시장에서 거래 가능하게 만든 예탁증서
- 미국 종목 목록 수집 시 미국 본토 기업과 별도로 분류하는 경우가 많다

### ETF / ETN / Closed-End Fund

- 주식처럼 거래되지만 개별 기업 보통주(Common Stock)와는 구조가 다름
- 종목 마스터 구축 시 자산 유형(type) 분리가 필요

## 실무 분류 관점

- 1차 분류: `Exchange` (`NYSE`, `NASDAQ`, `NYSE American`, `OTC`)
- 2차 분류: `Security Type` (`Common Stock`, `ETF`, `ADR`, `Preferred`)
- 3차 분류: `Listing Status` (`Active`, `Delisted`, `Suspended`)

## 오해하기 쉬운 포인트

- `NASDAQ`은 거래소 이름이자 내부 tier를 가진 시장 구조다
- `S&P 500`, `Dow Jones`, `Nasdaq-100`은 거래소가 아니라 지수(Index)다
- 미국 시장은 "2개 시장"으로 단순화하기 어렵고 다층 구조다
