
## Índice

- [Comandos](comandos.md)

---

## Http Api (postman app)

O botão abaixo irá adicionar uma coleção de requests http para o kubernetes dentro do app postman...

<a href="https://app.getpostman.com/run-collection/71ef69f32e87014ab9e1" target="_blank"><img src="https://run.pstmn.io/button.svg"></a>

E estes são os diferentes ambientes. É necessário fazer o download e adicionar um a um no postman...

- [Kubernetes test](../assets/postman-files/kubernetes_luizalabs_aws_tst.postman_environment)
- [Kubernetes dev](../assets/postman-files/kubernetes_luizalabs_aws_dev.postman_environment)
- [KUBERNETES PROD - CUIDADO](../assets/postman-files/kubernetes_luizalabs_aws_prod.postman_environment)

---

## Links

- Hello World Demo: [http://kubernetes.io/docs/hellonode/](http://kubernetes.io/docs/hellonode/)
- Deployments: [http://kubernetes.io/docs/user-guide/deployments/](http://kubernetes.io/docs/user-guide/deployments/)
- Docker to Kubernetes: [http://kubernetes.io/docs/user-guide/docker-cli-to-kubectl/](http://kubernetes.io/docs/user-guide/docker-cli-to-kubectl/)

## Instalando kubernetes

- Vagrant: [https://coreos.com/kubernetes/docs/latest/kubernetes-on-vagrant-single.html](https://coreos.com/kubernetes/docs/latest/kubernetes-on-vagrant-single.html)
- Kube Solo: [https://github.com/TheNewNormal/kube-solo-osx](https://github.com/TheNewNormal/kube-solo-osx)
- Versão com Docker (redspread): [https://github.com/redspread/localkube](https://github.com/redspread/localkube)

## Tools

- Gerenciador de pacotes para kubernetes: [https://helm.sh/](https://helm.sh/)
- Command line tool: [https://github.com/redspread/spread](https://github.com/redspread/spread)

## Api Swagger do kubernetes

- [http://kubernetes.io/kubernetes/third_party/swagger-ui/#!/apis/getAPIVersions](http://kubernetes.io/kubernetes/third_party/swagger-ui/#!/apis/getAPIVersions)





 
---

## ANOTAÇÕES GERAIS AINDA DESORGANIZADAS

`TODO: desconsiderar este. Anotações ainda precisam ser organizadas.`

```bash
+------------+  Cria/Gerencia  +--------------+  Cria/Gerencia  +------+
| Deployment +-----------------> Replica Sets +-----------------> Pods |
+------------+                 +--------------+                 +------+
```

Run

STEP 01  
- A: Comando executado
- B: Cria deployment do app com nome meu-app
- C: Cria replica set para gerenciar os pods.
  - Componente responsável por manter o que foi descrito no deployment do passo B. Ex.: Número de replicas
- D: Cria pods (4 pods por conta das 4 replicas)
```bash
$ kubectl run meu-app --image=gcr.io/meu-projeto-no-meu-registry-privado/meu-app:v1 --port=8080 --replicas=4
```

STEP 02  
- E: Cria o serviço responsável por rotear requests para os pods
- F: Cria rota externa no load balancer apontando para o serviço
```bash
$ kubectl expose deployment meu-app --type="LoadBalancer"
```

```bash
+---------------------------+----------------------------+
|                           |                            |
|   STEP 1                  |                 STEP 2     |
|                           +                            |
|   +-------------+                    +--------+        |
| A |$ Kubectl run|     D              |Internet|        |
|   +-----+-------+       +---+        +--+-----+        |
|         |  cria       +->Pod<-+         |              |
|   +-----v----+        | +---+ |      +--v----------+   |
| B |Deployment|        |       |      |Load Balancer| F |
|   +-----+----+        | +---+ |      +--+----------+   |
|         |             +->Pod<-+         |              |
|   +-----v-----+ cria  | +---+ | http +--v----+         |
| C |Replica Set+-------+       +------+Service| E       |
|   +-----------+       | +---+ |      +-------+         |
|                       +->Pod<-+                        |
|                       | +---+ |                        |
|                       |       |                        |
|                       | +---+ |                        |
|                       +->Pod<-+                        |
|                         +---+                          |
|                                                        |
|                           +                            |
+---------------------------+----------------------------+
```

`TODO:` Descrever melhor o processo de criação do serviço apontando para os pods.  
O serviço foi criado apontando para o deployment, por isso ele consegue direcionar tráfego para os 4 pods. Para isto ele precisa do label do deployment como selector. 