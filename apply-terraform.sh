#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INFRA_DIR="${SCRIPT_DIR}/infra"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    echo "Uso: $0 [--auto-approve]"
    echo ""
    echo "Opções:"
    echo "  --auto-approve         Aplica sem confirmação"
    echo ""
    echo "As variáveis devem ser fornecidas via:"
    echo "  - Arquivo .env na raiz do projeto"
    exit 1
}

# Carrega variáveis do .env se existir
if [ -f "${SCRIPT_DIR}/.env" ]; then
    set -a
    source "${SCRIPT_DIR}/.env"
    set +a
else 
    echo "Crie e preenche as variáveis do seu arquivo .env"
fi

LAB_ROLE_ARN="${AWS_TERRAFORM_ROLE:-}"
DATABASE_PASSWORD="${DB_PASSWORD:-}"
AUTO_APPROVE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --auto-approve)
            AUTO_APPROVE="-auto-approve"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Erro: Opção desconhecida: $1${NC}"
            usage
            ;;
    esac
done

if [ -z "$LAB_ROLE_ARN" ]; then
    echo -e "${YELLOW}lab_role_arn não encontrado.${NC}"
    exit 1
fi

if [ -z "$DATABASE_PASSWORD" ]; then
    echo -e "${YELLOW}database_password não encontrado.${NC}"
    exit 1
fi

# Exporta variáveis para o Terraform
export TF_VAR_lab_role_arn="$LAB_ROLE_ARN"
export TF_VAR_db_password="$DATABASE_PASSWORD"

echo -e "${GREEN}Iniciando Terraform...${NC}"
echo "Diretório: ${INFRA_DIR}"
echo ""

cd "${INFRA_DIR}"

echo -e "${GREEN}[1/3] Executando terraform init...${NC}"
terraform init

echo -e "${GREEN}[2/3] Executando terraform plan...${NC}"
terraform plan -out=tfplan

echo -e "${GREEN}[3/3] Executando terraform apply...${NC}"
if [ -n "$AUTO_APPROVE" ]; then
    terraform apply -auto-approve tfplan
else
    terraform apply tfplan
fi

echo -e "${GREEN}✅ Terraform apply concluído com sucesso!${NC}"
