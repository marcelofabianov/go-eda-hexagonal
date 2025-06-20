ENV_TEMPLATE_DIR := _env/dev
DOCKERFILE_TEMPLATE_NAME := Dockerfile
DOCKER_COMPOSE_TEMPLATE_NAME := docker-compose.yml
ENV_FILE_TEMPLATE_NAME := dev.env
ALIASES_TEMPLATE_NAME := .project_aliases.example.sh

DOCKERFILE_SOURCE := $(ENV_TEMPLATE_DIR)/$(DOCKERFILE_TEMPLATE_NAME)
DOCKER_COMPOSE_SOURCE := $(ENV_TEMPLATE_DIR)/$(DOCKER_COMPOSE_TEMPLATE_NAME)
ENV_FILE_SOURCE := $(ENV_TEMPLATE_DIR)/$(ENV_FILE_TEMPLATE_NAME)
ALIASES_SOURCE := $(ENV_TEMPLATE_DIR)/$(ALIASES_TEMPLATE_NAME)

ROOT_DOCKERFILE_TARGET := ./Dockerfile
ROOT_DOCKER_COMPOSE_TARGET := ./docker-compose.yml
ROOT_ENV_FILE_TARGET := ./.env
ROOT_ALIASES_TARGET := ./.project_aliases.sh

REQUIRED_TEMPLATES := $(DOCKERFILE_SOURCE) $(DOCKER_COMPOSE_SOURCE) $(ENV_FILE_SOURCE) $(ALIASES_SOURCE)
GENERATED_ROOT_FILES := $(ROOT_DOCKERFILE_TARGET) $(ROOT_DOCKER_COMPOSE_TARGET) $(ROOT_ENV_FILE_TARGET) $(ROOT_ALIASES_TARGET)

.PHONY: help default setup-dev clean-dev check-templates

default: help

help:
	@echo "---------------------------------------------------------------------------------"
	@echo " Comandos Disponíveis para Gerenciamento de Ambiente do Projeto"
	@echo "---------------------------------------------------------------------------------"
	@echo " make setup-dev     - Configura o ambiente de desenvolvimento na raiz do projeto."
	@echo "                      Copia Dockerfile, docker-compose.yml, processa .env,"
	@echo "                      e copia .project_aliases.sh de '$(ENV_TEMPLATE_DIR)'."
	@echo "                      HOST_UID e HOST_GID são preenchidos automaticamente no .env."
	@echo ""
	@echo " make clean-dev     - Remove os arquivos de ambiente gerados da raiz do projeto."
	@echo ""
	@echo " make check-templates - Verifica se todos os arquivos de template necessários existem"
	@echo "                        em '$(ENV_TEMPLATE_DIR)'."
	@echo "---------------------------------------------------------------------------------"

check-templates:
	@echo "INFO: Verificando arquivos de template em '$(ENV_TEMPLATE_DIR)'..."
	@missing_files=0; \
	for template_file in $(REQUIRED_TEMPLATES); do \
		if [ ! -f "$$template_file" ]; then \
			echo "ERRO: Template NÃO ENCONTRADO: $$template_file"; \
			missing_files=1; \
		fi; \
	done; \
	if [ $$missing_files -eq 1 ]; then \
		echo "      Crie os templates faltantes em '$(ENV_TEMPLATE_DIR)' e tente novamente."; \
		exit 1; \
	fi
	@echo "INFO: Todos os templates necessários foram encontrados."

setup-dev: check-templates
	@echo "INFO: Iniciando configuração do ambiente de desenvolvimento na raiz do projeto..."
	@echo "      - Criando diretórios temporários necessários..."
	@mkdir -p ./tmp/air
	@echo "      - Copiando '$(DOCKERFILE_SOURCE)' para '$(ROOT_DOCKERFILE_TARGET)'..."
	@cp "$(DOCKERFILE_SOURCE)" "$(ROOT_DOCKERFILE_TARGET)"
	@echo "      - Copiando '$(DOCKER_COMPOSE_SOURCE)' para '$(ROOT_DOCKER_COMPOSE_TARGET)'..."
	@cp "$(DOCKER_COMPOSE_SOURCE)" "$(ROOT_DOCKER_COMPOSE_TARGET)"
	@echo "      - Gerando '$(ROOT_ENV_FILE_TARGET)' a partir de '$(ENV_FILE_SOURCE)'..."
	@ > "$(ROOT_ENV_FILE_TARGET)"
	@echo "# Este arquivo .env foi gerado por 'make setup-dev' em $$(date)." >> "$(ROOT_ENV_FILE_TARGET)"
	@echo "# NÃO FAÇA COMMIT DESTE ARQUIVO. Edite o template em '$(ENV_FILE_SOURCE)'." >> "$(ROOT_ENV_FILE_TARGET)"
	@echo "" >> "$(ROOT_ENV_FILE_TARGET)"
	@CURRENT_UID=$$(id -u) ; \
	CURRENT_GID=$$(id -g) ; \
	sed -e "s/^HOST_UID=.*/HOST_UID=$${CURRENT_UID}/" \
	    -e "s/^HOST_GID=.*/HOST_GID=$${CURRENT_GID}/" \
	    -e "s/\$${HOST_UID}/$${CURRENT_UID}/g" \
	    -e "s/\$${HOST_GID}/$${CURRENT_GID}/g" \
	    "$(ENV_FILE_SOURCE)" >> "$(ROOT_ENV_FILE_TARGET)"
	@echo "        '$(ROOT_ENV_FILE_TARGET)' gerado com HOST_UID=$$(id -u) e HOST_GID=$$(id -g)."
	@echo "      - Copiando '$(ALIASES_SOURCE)' para '$(ROOT_ALIASES_TARGET)'..."
	@cp "$(ALIASES_SOURCE)" "$(ROOT_ALIASES_TARGET)"
	@echo "INFO: Configuração do ambiente de desenvolvimento concluída."
	@echo "      Os seguintes arquivos foram criados/atualizados na raiz do projeto:"
	@for root_file in $(GENERATED_ROOT_FILES); do \
		echo "        - $$root_file"; \
	done
	@echo "      Lembre-se de que esses arquivos devem estar listados no seu .gitignore."

clean-dev:
	@echo "INFO: Removendo arquivos de ambiente de desenvolvimento da raiz do projeto..."
	@echo "      Serão removidos: $(GENERATED_ROOT_FILES)"
	@rm -f $(GENERATED_ROOT_FILES)
	@echo "INFO: Arquivos de ambiente da raiz removidos."

