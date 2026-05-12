terraform {
  backend "s3" {
    bucket = "tech-challenge-13-soat-tfstate"
    key = "tfstate"
    region = "us-east-1"
  }
}

module "network" {
  source       = "./modules/network"
  cluster_name = var.cluster_name
  vpc_cidr     = var.vpc_cidr
}

module "eks" {
  source       = "./modules/eks"
  cluster_name = var.cluster_name

  # Pass the LabRole ARN down to the EKS module
  lab_role_arn = var.lab_role_arn

  # Pass AWS network details
  vpc_id          = module.network.vpc_id
  public_subnets  = module.network.public_subnets
  private_subnets = module.network.private_subnets
}

module "rds" {
  source = "./modules/rds"

  vpc_id          = module.network.vpc_id
  private_subnets = module.network.private_subnets

  eks_cluster_security_group_id = module.eks.cluster_security_group_id

  db_name = "garagedb"
  db_user = "postgres"
  db_pass = var.db_password

  depends_on = [module.eks]
}