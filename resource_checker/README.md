# Resource Checker

본 도구의 경우 NHN의 느린 Web UI와 인스턴스 이름 및 IP, SG 등을 복붙하기 힘들어 만들었습니다. 
CLI/UI를 통해 빠르게 NHN Cloud의 Component 정보를 가져오세요.
GET 요청만 해줍니다.




``` YAML
nhn:
  Resourcechecker:
    enabled: true
    mode: {cli|gui|both}

```

## config.yaml 찾는 순서

`--config`로 직접 지정하지 않으면 다음 순서로 찾는다:
1. 현재 디렉토리부터 부모 디렉토리로 거슬러 올라가며 `config.yaml` 탐색 (저장소 안에서
   `go run .`처럼 로컬 개발할 때 어느 하위 폴더에서 실행해도 자동으로 찾힘)
2. 위에서 못 찾으면 `~/.config/nhn_iac/config.yaml`을 기본값으로 사용한다 — 파일이 아직
   없어도 이 경로가 기본값이므로, `rescheck configure set`을 처음 실행하면(설치된
   `rescheck`를 cron이나 임의의 디렉토리에서 호출하는 경우 포함) 이 위치에 새로 생성된다
   (부모 디렉토리도 자동으로 만들어짐)

즉 최초 설정도, 이후 모든 조회도 기본적으로 `~/.config/nhn_iac/config.yaml` 하나만 보게 된다.





## CLI Mode

``` bash
## USAGE
rescheck [OPTIONS] [COMMAND]

set
  config 


compute
  ls
    -a all list
    -l 
  



network

help

```