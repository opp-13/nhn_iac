---
paths:
  - "auto_remove_resource/**"
---

# Auto Remove Resource 규칙

- 실제 리소스를 삭제(DELETE)하는 파괴적 코드다. 삭제 로직을 변경할 때는 dry-run 모드나 확인 프롬프트를 우선 고려한다
- `auto_remove_resource/nhn`은 `resource_checker` 패키지에 의존해 리소스 목록을 조회한 뒤 삭제한다 — 조회 로직을 이 안에서 직접 재구현하지 않는다
- `auto_remove_resource/aws`는 아직 미구현이다. 새 provider를 추가할 때는 `nhn` 패키지와 동일한 인터페이스를 따른다 (`add-provider` 스킬 참고)
