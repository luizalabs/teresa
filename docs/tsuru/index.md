# tsuru

Site: [https://tsuru.io](https://tsuru.io)  
Docs: [https://docs.tsuru.io](https://docs.tsuru.io)

## Nuvem Luiza

- Topologia de redes atual instalada na nuvemluiza: [topologia v3](topologia.md)
- Instalação: [instalação](instalacao.md)

### Administradores

- eder.prado@luizalabs.com
- tsuru-admin@luizalabs.com
- renato@luizalabs.com
- arnaldo@luizalabs.com

### Adicionando um novo nó ao cluster docker
```bash
# pool-teste é o nome do pool de recursos já previamente cadastrado
$ tsuru-admin docker-node-add --register address=http://10.152.118.74:2375 pool=pool-teste
```

### Adicionando usuário para um time específico
```bash
$ yes "123456" | tsuru user-create <user-email>
# se o time já existir, nada acontece
$ tsuru team-create <time>
$ tsuru role-assign team-member <user-email> <time>
# Até este ponto o usuário já foi criado e adicionado ao time
```

Orientar o usuário nos seguintes passos...
```bash
# Vai adicionar o novo target e já selecionar ele (o nome nuvemluiza é um alias, pode ser alterado para qualquer nome)
$ tsuru target-add nuvemluiza http://api-tsuru.nuvemluiza.intranet:8080 -s
# Login... ainda não possue login por social
$ tsuru login <user-email>
# Senha default
$ 123456
# Trocar a senha por favor
$ tsuru change-password
# Adicionar a key para ser utilizada com o git
$ tsuru key-add my-key ~/.ssh/id_rsa.pub
```

### Instalando plataformas
```bash
$ tsuru-admin platform-add nodejs -d https://raw.github.com/tsuru/basebuilder/master/nodejs/Dockerfile
$ tsuru-admin platform-add python -d https://raw.github.com/tsuru/basebuilder/master/python/Dockerfile
$ tsuru-admin platform-add python3 -d https://raw.github.com/tsuru/basebuilder/master/python3/Dockerfile
$ tsuru-admin platform-add go -d https://raw.github.com/tsuru/basebuilder/master/go/Dockerfile
```

### Instalando serviços

- [Mysql Compartilhado](servicos.md#mysql-compartilhado)





