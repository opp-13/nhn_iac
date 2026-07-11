---
paths:
  - "**/*.go"
---

# Go 코드 스타일

- credential(tenantId, username, password, 발급된 토큰)은 로그나 error 메시지에 절대 포함하지 않는다
- 설정은 하나의 `config.yaml`에서 `nhn.<모듈명>` 키 구조로 읽는다 (예: `nhn.Resourcechecker`, `nhn.Autoremover`) — 모듈별로 별도 config 파일을 만들지 않는다
- 클라우드 provider별 API 클라이언트는 패키지로 분리하고, 상위 모듈은 구체 타입이 아닌 인터페이스에 의존한다
