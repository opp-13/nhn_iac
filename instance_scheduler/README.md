# 인스턴스 스케줄러 (instsched)

인스턴스마다 설정한 시각(cron 스케줄)에 그 NHN Cloud 인스턴스가 존재/실행 중임을 보장한다.
- 정지됨(SHUTOFF) → `rescheck compute run`으로 재시작
- 삭제됨 → `terraform: true`로 설정한 인스턴스만 Terraform으로 재생성 (기존 VPC/subnet 재사용).
  `terraform: false`면 복구하지 않고 로그만 남기고 넘어간다.

이 모듈은 `resource_checker/`와 완전히 독립된 Go 모듈이다 (자체 `go.mod`/`go.sum`/`main.go`).
Go 코드를 공유하지 않고, 빌드된 `rescheck` 바이너리를 PATH에서 서브프로세스로 호출해
인스턴스 상태를 조회/시작한다. **`rescheck`가 먼저 PATH에 설치되어 있어야 한다**
(루트 `install.sh` 참고).

## 설정 (config.yaml)

`resource_checker`와 같은 `config.yaml` 파일의 `nhn.auth`(공유) + `nhn.Instancescheduler` 키를
읽는다. 인스턴스마다 개별 스케줄과 terraform 복구 가능 여부를 갖는다.

**config.yaml을 찾는 순서** (`--config`로 직접 지정하지 않은 경우):
1. 현재 디렉토리부터 부모 디렉토리로 거슬러 올라가며 `config.yaml` 탐색 (저장소 안에서
   `go run .`처럼 로컬 개발할 때 어느 하위 폴더에서 실행해도 자동으로 찾힘)
2. 위에서 못 찾으면 `~/.config/nhn_iac/config.yaml`을 기본값으로 사용한다 (`rescheck
   configure set`을 처음 실행할 때도 이 경로에 새로 생성됨 — `resource_checker/README.md`
   참고). 어느 디렉토리에서 `instsched`/`rescheck`를 실행하든(cron이 어떤 cwd로 실행하든)
   항상 같은 config를 찾는다.

```yaml
nhn:
  auth:
    tenantId: {YOUR_TENANT_ID}
    region: {REGION}
    username: {YOUR_USER_NAME}
    password: {YOUR_PASSWORD}

  Instancescheduler:
    enabled: true
    mode: cli               # gui/both는 미구현
    terraformDir: ./instance_scheduler/terraform
    instances:
      - name: my-web-01
        schedule: "0 9 * * *"        # 매일 09:00에는 있어야 함
        terraform: true               # 삭제됐어도 terraform apply로 복구 가능
      - name: my-worker-01
        schedule: "*/30 9-18 * * *"
        terraform: false              # 삭제됐으면 복구하지 않고 로그만 남김
```

`schedule`은 cron 표현식(`분 시 일 월 요일`)이다 — **범위가 아니라 특정 시점**을 뜻한다.
`"0 9 * * *"`는 "0시~9시 사이에 떠있어야 함"이 아니라 "매일 09:00 정각에 딱 한 번 점검한다"는
뜻이다. 특정 시간대 내내 떠있는지 계속 확인하려면 `"*/10 9-18 * * *"`(9~18시 사이 10분마다)처럼
범위와 반복 주기를 조합해야 한다.

`terraform: true`인 인스턴스 이름은 `terraform/terraform.tfvars`의 `instances` map key와
정확히 일치해야 한다 (같은 이름으로 클라우드 상태와 terraform 정의를 매칭한다).
`terraform: false`인 인스턴스는 tfvars에 넣을 필요가 없다.

`terraformDir`가 상대경로면 (기본값 `./instance_scheduler/terraform`처럼) **config.yaml이
있는 디렉토리 기준**으로 해석된다 (cron의 예측 불가능한 cwd에 의존하지 않기 위함). config.yaml을
`~/.config/nhn_iac/config.yaml`에 두면서 실제 저장소는 다른 곳(예: `~/nhn_iac/`)에 클론했다면,
상대경로는 맞지 않으니 `terraformDir`에 저장소의 실제 절대경로(예:
`/root/nhn_iac/instance_scheduler/terraform`)를 직접 적어야 한다.

인스턴스마다 NIC(port)를 `openstack_networking_port_v2`로 명시적으로 프로비저닝한다 —
`instances[].fixed_ip`를 지정하면 그 IP로 고정되고, 생략(`null`)하면 subnet의 DHCP 할당에
맡긴다.

## Terraform 초기 설정 (최초 1회, 수동 — terraform: true인 인스턴스가 있을 때만 필요)

```bash
cd instance_scheduler/terraform
cp terraform.tfvars.example terraform.tfvars   # 실제 값 채우기 — 커밋 금지
terraform init
```

`terraform apply -auto-approve`는 이후 `instsched check`가 삭제 감지 시 자동으로 실행한다
(인증은 config.yaml의 `nhn.auth` 값을 `OS_AUTH_URL`/`OS_USERNAME`/`OS_PASSWORD`/
`OS_TENANT_ID`/`OS_REGION_NAME` 환경변수로 주입 — clouds.yaml 등 별도 credential 파일은
쓰지 않는다).

## CLI 사용법

```bash
instsched check [--instance NAME] [--config PATH]   # 1회 점검+복구
instsched schedule add <NAME> [--config PATH]        # crontab에 등록/갱신
instsched schedule remove <NAME>                     # crontab에서 제거
instsched schedule list                              # 등록된 스케줄 목록
instsched -h / --help
instsched -v / --version
```

`check --instance NAME`은 그 인스턴스 하나만 점검한다 (`schedule add`가 등록하는 crontab
항목이 실제로 쓰는 형태). `--instance` 없이 실행하면 설정된 인스턴스 전체를 한 번에
점검한다 (수동 확인/테스트용).

`check`는 사람이 읽기 쉬운 결과를 출력한다:

```
my-web-01: OK
my-worker-01: STOPPED -> started
my-db-01: DELETED -> terraform apply 완료
my-batch-01: DELETED (terraform 미설정 — 복구하지 않음)
```

인스턴스별 조치가 하나라도 실패하면 exit code 1을 반환한다. `enabled: false`거나
`instances`가 비어 있으면 아무 조치 없이 exit code 0으로 종료한다.

## 스케줄 등록 (crontab)

`schedule add`/`schedule remove`가 현재 사용자의 crontab에 인스턴스별 항목을
등록/삭제한다 (다른 crontab 라인은 건드리지 않음 — 마커 주석으로 관리 항목만 식별):

```bash
instsched schedule add my-web-01     # config.yaml의 schedule 값으로 crontab에 등록
instsched schedule list              # 현재 등록된 항목 확인
instsched schedule remove my-web-01  # 제거
```

**crontab 권한**: 일반 사용자 권한으로 자신의 crontab을 등록/조회하는 것은 보통 문제
없지만, 실행 계정에 crontab 사용 권한이 없는 환경(일부 컨테이너/제한된 서버)에서는
`crontab` 명령 자체가 실패할 수 있다. 이 경우 `schedule add/remove`는 실패 메시지에
"sudo로 다시 시도해 보세요"라는 안내를 덧붙인다 — 정확한 권한 오류 판별이 아니라
실패 시 붙는 일반적인 힌트이므로, 실제로는 계정/crontab 설정을 먼저 확인하는 것이 좋다.

물론 `crontab -e`로 직접 등록/수정해도 된다 — `schedule add/remove`는 편의 기능일 뿐이다.

## 알려진 제한

- `ERROR`, `BUILD` 등 `ACTIVE`/`SHUTOFF` 외 상태는 자동 조치하지 않고 로그만 남긴다.
- `check --instance` 없이 전체 점검할 때, `terraform: true`인 인스턴스가 여러 개 동시에
  삭제됐어도 `terraform apply`는 한 번만 실행된다 (`for_each` 기반 설정이라 한 번의 apply로
  누락된 리소스가 모두 재생성됨).
