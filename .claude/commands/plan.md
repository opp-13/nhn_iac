---
description: 코드를 조사하고 구현 계획을 PLAN.md로 작성한다 (코드 수정 없음)
argument-hint: <구현하려는 기능/변경 사항>
---

다음 작업에 대한 구현 계획을 세운다: $ARGUMENTS

1. **조사** — 관련 모듈(`resource_checker`, `auto_remove_resource/nhn`, `auto_remove_resource/aws`)과 `CLAUDE.md`, `.claude/rules/`를 읽고 현재 구조·제약을 파악한다. 이 단계에서는 코드를 수정하지 않는다.
2. **계획 작성** — 변경할 파일, 순서, 각 단계의 검증 방법(`go build ./...`, `go vet ./...`, `go test ./...`)을 포함한 계획을 `PLAN.md`에 작성한다.
   - 새 provider 추가라면 `.claude/skills/add-provider` 스킬의 단계를 반영한다.
   - 삭제/변경성 동작이 포함되면 dry-run 또는 확인 절차를 계획에 명시한다.
3. 계획 작성 후 **사용자 승인 전까지 구현을 시작하지 않는다.**
