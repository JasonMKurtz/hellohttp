# A small, public-only VPC: no NAT gateway, so it costs nothing while idle.
# Nodes get public IPs and reach the internet (ECR, etc.) through the IGW.
#
# Dual-stack, for an IPv6 cluster (eks.tf): pods and the load balancer use
# IPv6, while nodes also keep an IPv4 address. EKS needs subnets with both;
# with no NAT, the nodes' public IPv4 is also how they reach IPv4-only
# services.

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, 3)
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  # AWS assigns the VPC a /56 of public IPv6 (you don't pick it; see the
  # vpc_ipv6_cidr output), split into a /64 per subnet below.
  assign_generated_ipv6_cidr_block = true

  tags = { Name = var.cluster_name }
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = { Name = var.cluster_name }
}

resource "aws_subnet" "public" {
  count = length(local.azs)

  vpc_id                  = aws_vpc.this.id
  availability_zone       = local.azs[count.index]
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  map_public_ip_on_launch = true

  # One /64 of the VPC's /56 per subnet (8 more prefix bits). Every network
  # interface launched here gets an IPv6 address, and pods get theirs from
  # this range too.
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.this.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = "${var.cluster_name}-public-${local.azs[count.index]}"
    # Lets EKS place internet-facing load balancers here.
    "kubernetes.io/role/elb" = "1"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  # IPv6 needs its own default route. The same internet gateway carries it;
  # IPv6 addresses are public, so there's no NAT either way.
  route {
    ipv6_cidr_block = "::/0"
    gateway_id      = aws_internet_gateway.this.id
  }

  tags = { Name = "${var.cluster_name}-public" }
}

resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}
