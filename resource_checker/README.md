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