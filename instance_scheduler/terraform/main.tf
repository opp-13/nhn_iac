terraform {
  required_version = ">= 1.3.0" # optional object attributes (instances[].fixed_ip)

  required_providers {
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

# 인증은 OS_AUTH_URL / OS_USERNAME / OS_PASSWORD / OS_TENANT_ID / OS_REGION_NAME
# 환경변수로 주입된다 (instance_scheduler/guard가 config.yaml의 nhn.auth 값으로
# 채워 넣음) — credential을 이 저장소나 tfvars에 두지 않는다.
provider "openstack" {}

# 기존 VPC/subnet을 재사용한다 (신규 network/subnet 리소스는 만들지 않음).
# terraform: false로 설정한 인스턴스는 여기 instances map에 넣을 필요가 없다
# (삭제돼도 이 프로젝트가 복구하지 않으므로).

# 인스턴스마다 NIC(port)를 명시적으로 프로비저닝한다 — fixed_ip를 지정하면 그 IP로
# 고정되고, 생략하면 subnet의 DHCP 할당에 맡긴다.
resource "openstack_networking_port_v2" "this" {
  for_each = var.instances

  name       = "${each.key}-port"
  network_id = var.network_id

  dynamic "fixed_ip" {
    for_each = each.value.fixed_ip != null ? [each.value.fixed_ip] : []
    content {
      subnet_id  = var.subnet_id
      ip_address = fixed_ip.value
    }
  }
}

resource "openstack_compute_instance_v2" "this" {
  for_each = var.instances

  name            = each.key
  flavor_id       = each.value.flavor_id
  image_id        = each.value.image_id
  key_pair        = each.value.key_pair
  security_groups = each.value.security_groups

  network {
    port = openstack_networking_port_v2.this[each.key].id
  }
}
