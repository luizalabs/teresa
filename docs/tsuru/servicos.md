# Serviços do tsuru

## Mysql Compartilhado

Projeto: [https://github.com/tsuru/mysqlapi](https://github.com/tsuru/mysqlapi)

**Preparando o mysql compartilhado**   

Antes de mais nada é necessário um mysql (ou mariadb) previsamente instalado.  

```bash
# Entrar no mysql com o usuário root...
$ mysql -h <ip-do-mysql> -u root -p
> GRANT ALL PRIVILEGES ON *.* TO ‘<tsuru>'@'%' IDENTIFIED BY ‘<senha-db-aqui>' with GRANT OPTION;
> FLUSH PRIVILEGES;

# Entrar no mysql com o usuário ’tsuru’ recém criado para criar a base de dados que será utilizada pela api
$ mysql -h <ip-do-mysql> -u tsuru -p
> CREATE DATABASE <mysqlapi>;
```

**Criando o app responsável pelo controle do serviço...**  

```bash
# Logar com um usuário com acesso admin no tsuru...
$ tsuru login
# Baixar o projeto da api-mysql do tsurru...
$ git clone git@github.com:tsuru/mysqlapi.git tsuru-mysqlapi
$ cd tsuru-mysqlapi/
# Verificar se existe um time para adicionar os projetos, uma vez que a API será executada como um app do tsuru...
$ tsuru team-list
Se não existir, criar o time...
$ tsuru team-create tsuru-services

# Criando a aplicação que fara o controle do serviço...
$ tsuru app-create tsuru-service-mysql-api python -t tsuru-services

# Adicionando variáveis de ambiente no app...
$ tsuru env-set -a tsuru-service-mysql-api DJANGO_SETTINGS_MODULE=mysqlapi.settings
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_DB_NAME=<mysqlapi>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_DB_USER=<tsuru>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_DB_PASSWORD=<senha-db-aqui>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_DB_HOST=<ip-do-mysql>

# Recuperando o nome do repositório do tsuru...
$ tsuru app-info -a tsuru-service-mysql-api | grep Repository
# Deploy do app...
$ git push <git@git-tsuru.nuvemluiza.intranet:tsuru-service-mysql-api.git> master

# Adicionando variáveis de ambiente para utilizar o mysql compartilhado
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_SHARED_SERVER=<ip-do-mysql>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_SHARED_SERVER_PUBLIC_HOST=<ip-do-mysql>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_SHARED_USER=<tsuru>
$ tsuru env-set -a tsuru-service-mysql-api MYSQLAPI_SHARED_PASSWORD=<senha-db-aqui>

# Neste ponto a api do serviço deve estar rodando no tsuru
```

**Adicionando o novo serviço**  
```bash
# Na pasta do projeto, iniciar o template para instalar o serviço...
$ crane template

# Alterar o arquivo manifest.yaml seguindo o exemplo
#   O endpoint dentro do manifest é o endpoint do app no tsuru
id: mysql
username: <tsuru>
password: <senha-db-aqui>
team: <tsuru-services>
endpoint:
  production: tsuru-service-mysql-api.sandbox.nuvemluiza.intranet

# Adicionando o serviço a partir do manifest.yaml
$ crane create ./manifest.yaml
# Neste ponto o serviço já deve estar funcionando e disponível!!!
```

**Testando o serviço**  
```bash
$ tsuru service-list
# O serviço mysql deverá aparecer na lista

# Criando uma nova bada de dados
$ tsuru service-add mysql <nome-da-base-mysql-para-app>
# Fazendo o bind da base de dados com o app
$ tsuru service-bind mysql <nome-da-base-mysql-para-app> -a <nome-do-app>

# Pronto, neste ponto as variáveis de ambiente para acesso no serviço serão exibidas no shell e já estarão disponíveis para o app que foi feito o bind
```

## DIAATS

Projeto: [https://github.com/tsuru/diaats](https://github.com/tsuru/diaats)

