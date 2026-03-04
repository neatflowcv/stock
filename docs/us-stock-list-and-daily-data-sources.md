# 미국 주식 목록 + 하루 단위 조회 가능 소스 조사

- 조사일: 2026-03-04
- 목적: `미국 주식 종목 목록(티커 마스터)` 조회와 `하루 단위(일봉/특정 날짜 기준)` 조회가 가능한 서비스 확인

## 결론 요약

가장 실무적으로는 아래 조합이 안정적입니다.

1. **Nasdaq Trader Symbol Directory**: 거래소 기준 심볼 마스터(파일) 확보
2. **Alpha Vantage LISTING_STATUS(date)**: 특정 날짜 기준 상장 목록(as-of) 조회
3. **Polygon / Twelve Data / FMP**: 일봉(EOD) 시세 조회

## 후보별 정리

- Nasdaq Trader Symbol Directory
  - 미국 종목 목록 조회: 가능 (파일 다운로드)
  - 하루 단위 목록/데이터 조회: 가능 (Daily List로 일별 변경 추적)
  - 비고: 거래소 원천에 가까운 심볼 파일
- Alpha Vantage
  - 미국 종목 목록 조회: 가능 (`LISTING_STATUS`)
  - 하루 단위 목록/데이터 조회: 가능 (`date=YYYY-MM-DD`, `TIME_SERIES_DAILY`)
  - 비고: 간단한 HTTP API, 키 필요
- Polygon.io
  - 미국 종목 목록 조회: 가능 (`/v3/reference/tickers`)
  - 하루 단위 목록/데이터 조회: 가능 (일봉 aggregates, 날짜 파라미터)
  - 비고: 실시간/스냅샷 등 확장성 큼
- Twelve Data
  - 미국 종목 목록 조회: 가능 (`/stocks`)
  - 하루 단위 목록/데이터 조회: 가능 (`/time_series?interval=1day`)
  - 비고: 멀티마켓 API, 키 필요
- Financial Modeling Prep (FMP)
  - 미국 종목 목록 조회: 가능 (`/api/v3/stock/list`)
  - 하루 단위 목록/데이터 조회: 가능 (`historical-price-eod/full`)
  - 비고: 학습/프로토타입에 편함
- SEC company_tickers.json
  - 미국 종목 목록 조회: 부분 가능 (SEC 등록사 목록)
  - 하루 단위 목록/데이터 조회: 일봉 데이터 미제공
  - 비고: 거래 가능 전체 종목 마스터 대체재로는 부적합

## 상세

### 1) Nasdaq Trader Symbol Directory (우선 추천: 종목 마스터)

- 용도
  - 미국 거래소 관련 심볼 마스터 파일 확보
  - 거래소 제공 Daily List로 일별 추가/삭제 변경 추적
- 장점
  - 데이터 소스 신뢰도가 높음
  - 파일 기반이라 배치 수집에 적합
- 주의
  - API JSON이 아니라 파일 수집/파싱 워크플로우 필요

참고:

- Symbol Directory 정의: <https://www.nasdaqtrader.com/Trader.aspx?id=SymbolDirDefs>
- Daily List: <https://www.nasdaqtrader.com/Trader.aspx?id=DailyList>

### 2) Alpha Vantage (as-of 날짜 기준 상장 목록이 강점)

- 종목 목록
  - `function=LISTING_STATUS`
  - `date=YYYY-MM-DD` 파라미터로 특정 날짜 기준 상장 목록 조회 가능
- 하루 단위
  - `TIME_SERIES_DAILY`로 일봉 조회 가능
- 예시
  - `https://www.alphavantage.co/query?function=LISTING_STATUS&date=2025-12-31&apikey=YOUR_KEY`

참고:

- 공식 문서: <https://www.alphavantage.co/documentation/>

### 3) Polygon.io (대규모/실무 확장)

- 종목 목록
  - `GET /v3/reference/tickers`
- 하루 단위
  - Aggregates(일봉) API에서 날짜 범위/일 단위 조회 가능
- 비고
  - 스냅샷, 마켓 상태 등 부가 기능이 풍부

참고:

- Tickers: <https://polygon.io/docs/rest/stocks/tickers/all-tickers>
- Aggregates: <https://polygon.io/docs/rest/stocks/aggregates>

### 4) Twelve Data (단순 통합 API)

- 종목 목록
  - `GET /stocks`
- 하루 단위
  - `GET /time_series?symbol=AAPL&interval=1day`

참고:

- Stocks: <https://twelvedata.com/docs#stocks>
- Time Series: <https://twelvedata.com/docs#time-series>

### 5) Financial Modeling Prep (FMP)

- 종목 목록
  - `/api/v3/stock/list`
- 하루 단위
  - `/api/v3/historical-price-eod/full?symbol=AAPL`

참고:

- 문서/엔드포인트 예시: <https://site.financialmodelingprep.com/developer/docs/stable>

### 6) SEC company_tickers.json (보조용)

- 가능
  - CIK-티커-회사명 매핑 확보
- 한계
  - 거래소 기준의 “현재 거래 가능 전체 종목 마스터” 대체재로는 한계
  - 일봉/일별 시장 데이터 미제공

참고:

- 데이터 파일: <https://www.sec.gov/files/company_tickers.json>

## 실무 적용 권장안

1. **종목 마스터 원본**: Nasdaq Trader Symbol Directory
2. **특정 날짜 기준 목록(as-of) 보강**: Alpha Vantage `LISTING_STATUS(date)`
3. **일봉 시세**: Polygon 또는 Twelve Data 또는 FMP 중 비용/호출 제한에 맞춰 선택

## 최소 검증 체크리스트

- 심볼 중복/클래스주(`BRK.B` vs 포맷 차이) 정규화
- 상장폐지/티커변경 반영 주기 정의(일배치)
- API rate limit 초과 시 재시도 정책(백오프)
- 거래일 캘린더 기준으로 일봉 누락 처리
