# NHN 관리 도구 모음

NHN cloud 관리를 관리하기 위한 도구 모음 repo 입니다.
기본적으로 Openstack IaaS용 API credential로 인증하며 일부 서비스의 경우 user api로 인증합니다.

## Prerequisite

NHN Cloud [Openstack token]

Set your Credential
``` YAML
nhn:
  auth:
    tenantId: {YOUR_TENANT_ID}
    passwordCredentials:
      username: {YOUR_USER_NAME}
      password: {YOUR_PASSWORD}
```

