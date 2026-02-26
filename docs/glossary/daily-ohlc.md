# DailyOHLC

## 정의
`DailyOHLC`는 특정 `Stock`의 하루 가격 요약 데이터다.

## 필드
- `date`: 거래일
- `open`: 해당 거래일의 시가
- `high`: 해당 거래일의 최고가
- `low`: 해당 거래일의 최저가
- `close`: 해당 거래일의 종가

## 관계
- 하나의 `Stock`은 여러 `DailyOHLC`를 가진다.
- 하나의 `DailyOHLC`는 하나의 `Stock`에 속한다.
