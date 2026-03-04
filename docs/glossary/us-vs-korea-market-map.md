# USVsKoreaMarketMap

## 정의

`USVsKoreaMarketMap`은 한국의 `KOSPI/KOSDAQ` 개념을
미국 시장 구조에 대응시켜 설명하기 위한 용어 매핑이다.

## 용어 매핑

### KOSPI

- 한국의 유가증권시장
- 대형/전통 기업 비중이 상대적으로 높은 편

미국에서의 대응 개념(완전 일치 아님):

- `NYSE` 중심 대형주 상장 시장 이미지
- `NASDAQ Global Select Market`의 대형 우량주 구간

### KOSDAQ

- 한국의 코스닥시장
- 성장주/기술주 비중이 상대적으로 높은 편

미국에서의 대응 개념(완전 일치 아님):

- `NASDAQ` 전체 이미지(특히 기술주 중심 인식)
- `NASDAQ Capital Market`의 상대적 중소형 구간

### 핵심 차이

- 한국: `KOSPI/KOSDAQ`라는 제도적 2시장 구분이 명확
- 미국: 거래소(`NYSE`, `NASDAQ`)와 내부 tier,
  그리고 `OTC`가 함께 존재하는 다층 구조

## 미국 시장을 한국식으로 단순 비교할 때의 주의

- `NYSE=KOSPI`, `NASDAQ=KOSDAQ`은 교육용 비유로만 유효
- 실제 상장요건, 산업 분포, 유동성 구조는 직접 1:1 대응되지 않음
- 데이터 파이프라인에서는 반드시 `exchange`, `type`, `status`를
  별도 컬럼으로 관리해야 한다

## 데이터 모델링 권장 필드

- `symbol`: 티커
- `name`: 종목명
- `exchange`: 주 상장 거래소
- `market_tier`: 거래소 내부 tier (있을 때만)
- `security_type`: 보통주/ETF/ADR 등
- `listing_status`: 상장/정지/상폐
- `is_otc`: OTC 여부

## 한 줄 정리

미국 주식도 "나뉘긴" 하지만,
한국처럼 `코스피/코스닥` 2분할이 아니라
`거래소 + tier + OTC + 종목유형`으로 나뉜다.
