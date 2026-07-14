variable "network_id" {
  description = "재사용할 기존 NHN Cloud network(VPC) ID. 인스턴스별 NIC(port)가 이 네트워크에 붙는다. VPC/network은 이 프로젝트가 만들지 않는다."
  type        = string
}

variable "subnet_id" {
  description = "재사용할 기존 NHN Cloud subnet ID. instances[].fixed_ip을 지정한 경우 그 IP를 이 subnet 안에 고정하는 데 쓰인다. subnet은 이 프로젝트가 만들지 않는다."
  type        = string
}

variable "instances" {
  description = "terraform: true로 설정한 인스턴스 정의. map의 key는 config.yaml의 nhn.Instancescheduler.instances에 나열한 이름과 반드시 일치해야 instsched가 상태를 대응시킬 수 있다."
  type = map(object({
    flavor_id       = string
    image_id        = string
    key_pair        = string
    security_groups = list(string)
    fixed_ip        = optional(string) # 지정하면 이 IP로 NIC을 고정, 생략하면 subnet DHCP가 할당
  }))
}
