-include .env
export

CLUSTER_NAME=tech-challenge-api-local
IMAGE_NAME=tech-challenge-api:latest
AWS_ECR_IMAGE=$(AWS_ACCOUNT).dkr.ecr.$(AWS_REGION).amazonaws.com/$(AWS_ECR_REPO)

.PHONY: all up down deploy deploy-local deploy-aws switch-eck-aw build-aws apply-aws build load create-tfstate-bucket apply-terraform

up: build load deploy-local
	@echo "Waiting for tech-challenge-api deployment..."
	@sleep 3
	@kubectl rollout status deployment/tech-challenge-api --timeout=120s || (echo "Deployment não ficou disponível. Verificando status...";  exit 1)
	@kubectl wait --for=condition=ready pod -l app=tech-challenge-api --timeout=60s || (echo "Pods não ficaram prontos. Verificando status..."; exit 1)
	@echo "✅ App is running at http://localhost/"

build:
	docker build -t $(IMAGE_NAME) .

load:
	kind load docker-image $(IMAGE_NAME) --name $(CLUSTER_NAME)

deploy-local:
	@echo "Deploying local..."
	@COMPOSE_HOST=$$(ip addr show eth0 | grep 'inet ' | awk '{print $$2}' | cut -d/ -f1) && \
	echo "Docker host IP: $$COMPOSE_HOST" && \
	kubectl kustomize k8s/overlays/local | \
	  sed "s|COMPOSE_HOST|$$COMPOSE_HOST|g" | \
	  kubectl apply -f -

deploy-aws: apply-terraform switch-eck-aws deps build-aws apply-aws

switch-eck-aws:
	@echo "AWS CLI is needed and you should have been logged into your account ..."
	aws eks update-kubeconfig --region $(AWS_REGION) --name $(K8S_API_CLUSTER_NAME)

build-aws:
	@echo "Pushing docker image into AWS..."
	@echo "Docker and AWS CLI are needed before push API image..."
	docker build --target production -t $(AWS_ECR_IMAGE):latest .
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin ${AWS_ACCOUNT}.dkr.ecr.${AWS_REGION}.amazonaws.com
	docker push $(AWS_ECR_IMAGE):latest

apply-aws:
	@RDS_ADDRESS=$$(aws rds describe-db-instances --db-instance-identifier db-$(DB_NAME) --query="DBInstances[0].Endpoint.Address" --output text 2>/dev/null) && \
	if [ -z "$$RDS_ADDRESS" ]; then \
		echo "Erro: Não foi possível obter o endereço do RDS. Verifique se o Terraform foi aplicado."; \
		exit 1; \
	fi && \
	kubectl kustomize k8s/overlays/aws | \
	  sed "s|ECR_IMAGE:latest|$(AWS_ECR_IMAGE):latest|g" | \
	  kubectl apply -f - && \
	kubectl patch configmap api-config \
		--type merge \
		-p '{"data":{"DB_HOST":"'$$RDS_ADDRESS'"}}' && \
	kubectl rollout restart deployment tech-challenge-api && \
	kubectl get ingress tech-challenge-api-ingress -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'


deploy:
ifeq ($(ENV),aws)
	$(MAKE) deploy-aws
else
	$(MAKE) deploy-local
endif

create-tfstate-bucket:
	aws s3api create-bucket --bucket tech-challenge-13-soat-tfstate --region $(AWS_REGION)
	
	aws s3api put-bucket-versioning --bucket tech-challenge-13-soat-tfstate --versioning-configuration Status=Enabled

	aws s3api put-public-access-block  --bucket tech-challenge-13-soat-tfstate  --public-access-block-configuration \
	"BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

apply-terraform:
	@./apply-terraform.sh --auto-approve

down-terraform:
	export TF_VAR_lab_role_arn="$(AWS_TERRAFORM_ROLE)" && \
	export TF_VAR_db_password="$(DB_PASSWORD)" && \
	cd infra/ && terraform destroy -auto-approve