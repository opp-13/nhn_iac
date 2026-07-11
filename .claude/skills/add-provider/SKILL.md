---
name: add-provider
description: auto_remove_resource에 새로운 클라우드 provider(예 AWS) 지원을 추가하는 워크플로우
disable-model-invocation: true
---

새 클라우드 provider `$ARGUMENTS` 지원을 `auto_remove_resource`에 추가한다.

1. `auto_remove_resource/<provider>/` 디렉토리를 확인한다 (이미 scaffold되어 있으면 재사용)
2. `auto_remove_resource/nhn/`의 패키지 구조와 인터페이스를 참고해 동일한 방식으로 구현한다
3. `<provider>/README.md`를 `nhn/README.md`와 같은 형식(동작 방식 + 설정 스키마)으로 작성한다
4. `auto_remove_resource/README.md`의 지원 클라우드 표에서 상태를 "지원 예정" → "지원"으로 갱신한다
5. `go build ./...`, `go vet ./...`로 빌드를 확인한다
