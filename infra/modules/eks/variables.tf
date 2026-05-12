variable "cluster_name" {
  type = string
}

variable "lab_role_arn" {
  description = "The ARN of the LabRole provided by the FIAP student environment"
  type        = string
}

variable "vpc_id" {
  type = string
}

variable "public_subnets" {
  type = list(string)
}

variable "private_subnets" {
  type = list(string)
}