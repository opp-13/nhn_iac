# CLAUDE.md

NHN Cloud 관리 도구 모음 (모듈형 구조). 초기 개발 단계.

## 명령어

각 모듈은 독립된 Go 모듈이다 (자체 `go.mod`/`go.sum`/`main.go`) — 반드시 해당 폴더에서 빌드한다.

```bash
cd resource_checker && go build ./...        # rescheck 바이너리 빌드
cd resource_checker && go run . <args>       # 실행
cd resource_checker && go test ./...         # 테스트

cd instance_scheduler && go build ./...      # instsched 바이너리 빌드
cd instance_scheduler && go test ./...       # 테스트
```

## 모듈 구조

- `resource_checker/` — NHN Cloud 리소스 조회 (읽기 전용, GET 요청만 허용). `rescheck` 바이너리.
- `auto_remove_resource/` — 리소스 자동 정리
  - `nhn/` — NHN Cloud 구현 (현재 대상)
  - `aws/` — AWS 지원 예정 (미구현)
- `instance_scheduler/` — 인스턴스별로 설정한 시각(cron 스케줄)에 그 인스턴스가 존재/실행
  중임을 보장 (정지면 시작, 삭제면 terraform 설정된 경우에만 재생성). `instsched` 바이너리.

새 클라우드 지원은 `auto_remove_resource/<provider>/` 형태로 추가한다.

모듈 간에는 Go 패키지를 import하지 않는다 — 완전히 독립적으로 빌드/실행되며, 서로 필요하면
빌드된 CLI 바이너리를 서브프로세스로 호출한다 (예: `instance_scheduler`는 PATH의 `rescheck`를
호출). 새 provider/모듈을 추가할 때도 이 원칙을 따른다.

## 설정 규칙

- 모든 모듈은 하나의 `config.yaml`에서 `nhn.<모듈명>` 키로 설정을 읽는다 (스키마는 각 모듈 README 참고)
- 각 모듈은 `enabled: bool`, `mode: cli|gui|both` 공통 키를 가진다
- 인증: OpenStack IaaS API credential 기본, 일부 서비스는 User API 사용
- credential(tenantId, username, password)은 절대 커밋하지 않는다 — `.env`, 로컬 config만 사용
- 로컬 config는 절대 접근하지 않는다.

## 주의사항

- Autoremover는 Resourcechecker에 의존한다 (리소스 조회 후 제거)
- 문서와 커밋 메시지는 한국어 사용
- 모듈별 세부 코딩 규칙(GET 전용, 파괴적 동작 주의 등)은 `.claude/rules/`에 경로 기반으로 분리되어 있어 해당 경로 작업 시 자동 적용된다

## 문서 · 확장

- 각 모듈의 사용법·설정 스키마는 해당 디렉토리의 README.md 참고 (이 파일에 중복 기재하지 않음)
- 새 클라우드 provider 추가 워크플로우는 `.claude/skills/add-provider` 스킬 참고

## 개발 워크플로우

기능 개발은 `/plan` → `/execute` → `/ship` 순서로 진행한다 (`.claude/commands/`).
간단한 오타 수정 등 범위가 명확한 작업은 바로 진행해도 된다.
