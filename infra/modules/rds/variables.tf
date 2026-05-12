variable "db_name" {
  description = "Name of database"
  type        = string
  default     = "garagedb"
}

variable "engine_version" {
  description = "Postgres engine version"
  type        = string
  default     = "15"
}

variable "db_user" {
  description = "DB User name"
  type        = string
  default     = "postgres"
  sensitive = true
}

variable "db_pass" {
  description = "DB User pass"
  type        = string
  sensitive = true
}

variable "vpc_id" {
  description = "ID da VPC"
  type        = string
}

variable "private_subnets" {
  description = "IPs da subnet privada"
  type        = list(string)
}

variable "eks_cluster_security_group_id" {
  description = "ID do grupo de segurança"
  type        = string
}

variable "instance_class" {
  description = "Instância do RDS"
  type        = string
  default     = "db.t3.micro"
}

variable "allocated_storage" {
  description = "Espaço alocado para o DB"
  type        = number
  default     = 20
}
