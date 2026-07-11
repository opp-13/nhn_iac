---
paths:
  - "resource_checker/**"
---

# Resource Checker 규칙

- GET(조회) 또는 isntance를 shutwon 및 run하는 부분만 작성한다.
- 인증 실패(401/403) 등은 조용히 무시하지 말고 에러로 반환한다
- 출력 필드(인스턴스명, IP, SG 등)는 CLI에서 복붙하기 쉬운 평문/CSV 형태를 우선 고려한다
- github.com/rackspace/gophercloud 라이브러리를 사용한다. (https://pkg.go.dev/github.com/gophercloud/gophercloud/v2 참조)
- 만약 원하는 기능이 없을 경우 nhn의 https://docs.nhncloud.com/ko/nhncloud/ko/public-api/service-api/를 참조하여 제작한다.

