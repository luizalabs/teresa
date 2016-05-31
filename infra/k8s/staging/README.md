## Instalação do cluster do Kubernetes Desenvolvimento.##

**k8s-lab-env.sh** - Variáveis de ambiente para configurar o cluster kubernetes na AWS na Região US-EAST-1a com 2 nodes. É possível alterar o número de nodes depois do cluster criado, para isso, edite o ASG(AutoScaling Group) gerado automaticamente com o nome de **kubernetes-minion-group-us-east-1a**.

```
export KUBE_AWS_ZONE=us-east-1a
export AWS_S3_REGION=us-east-1
export NUM_NODES=2
export MASTER_SIZE=m3.medium
export NODE_SIZE=m3.large
export KUBE_ENABLE_NODE_PUBLIC_IP=true
export KUBE_ENABLE_CLUSTER_REGISTRY=true
export KUBERNETES_PROVIDER=aws
export KUBE_ENABLE_NODE_AUTOSCALER=true

```

    
No script `k8s-lab-env.sh` a flag `KUBE_ENABLE_CLUSTER_REGISTRY=true` habilita o docker registry no cluster, mas para subir o POD do mesmo, é necessário criar um persistent volume e o persistent volume claim. 
Na AWS usamos um *volume EBS*, falaremos mais no capítulo sobre configurações, após a criação do cluster.

----------

**k8s-lab-install.sh** - Este script carrega as variaveis de ambiente do script k8s-lab-env.sh (`source k8s-lab-env.sh`) e dispara o instalador do kubernetes.

Este processo irá executar algumas etapas, descritas abaixo:

1. Upload installation files to S3
1. Create IAM roles
1. Create a key pair and publish to AWS
1. Create VPC *
1. Create Subnet *
1. Create Internet Gateway
1. Create Routing Table
1. Create Security Groups
1. Create and attach persistent volume to master
1. Create master instance
1. Create node instances

Mandatoriamente o Kubernetes usa um VPC com bloco 172.20.0.0/16 e um SUBNET 172.20.0.0/24, caso você já tenha criado estes recursos, você precisará exportar as variáveis VPC_ID e SUBNET_ID com os ids dos mesmos.

Ao final da execução, será exibido o endereço do Master e seus serviços. 

```
Kubernetes master is running at https://52.202.226.173
Elasticsearch is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/kube-dns
KubeRegistry is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/kube-registry
kubernetes-dashboard is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
Grafana is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
InfluxDB is running at https://52.202.226.173/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb

Installation successful!
```

Foi criado uma chave ssh para conectar no master e nos nodes. Ela fica armazenada no diretório .ssh do HOME  do user que executou o `k8s-lab-install.sh`.
O nome gerado é `~/.ssh/kube_aws_rsa `
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAxB7f47jwx6TOX76bnmiOxxyVlu46CauyCf2RMilQ54W45nkU
OHKEMb827nN5ZmlgD3U9VgzDeqai9viJkoSb3UEN4c6s1AMBCzKHZ47XCw1eAiVL
AlfehjkIvGfzgMvm1JOKPNcF2nKyVPMOOA9OoDU2llXu0tjWjBKF975IqMG3XFlH
Rp8IfpHrGWoJC6tJQcuDm73QoGDfayqVOWwgxPGzONO/x6lZ9twyK8r9xu5xW3Y1
uAlcw2XaqprB52g27OnT2MNqwzTj3CLJvCqcwpNByRXQho3W5yxO6bXLZPRYaN47
yasMwHQlck8GuxqvAQoOJQ+h/vMjthCe43lU5QIDAQABAoIBACuWySAt7KYSxSOW
k7PjfUhX3A+NpyybEpkR2IGTmT28BNqEsq/RE/ySiTz6QVM3vHfzEMzdTV0nSDUo
DDuaaaIFYXZ8nwEIIeVBN8MWhkkYkXpcm41LxlOCvaRSXlaT+xvoJuNazxw8PdKP
qutey5Tj/tvvAYOvWhAL1ea4TiHjeIg7RCOappAEive7tRwNBGjvYqLKZTny66JC
tvWbMRwB+4Xe1e/i+PLrK3NhB2XZeKcwtifu4MFJ+p0/WoW3Bw6kAiqHJP0pbpsT
02nY0O5AzTLttF1VXi/WAoZrRVdLedKJkluAMStEWXzzyvp/k+BZi9OwRMaJgPet
DN7rhsECgYEA675TJEnGhQgBb4nhJP2jhigh09/PmGTbMbpMl5aNXzTM8vrG5J1P
LuRcmovZo5aTzmoRv7ydLp9T78+bc7k8t9tRYtY2oD9VGe+VO/mBmnxEhImit6jW
BEiGPetAj3kFce0afTVAk+mZ6WtwSbC7TyqVJewiTzJSvKe8xwJA5NkCgYEA1Pj2
IwHtvc0MysjjY1+lrz27DusMjdpLMhTdIC7xDo5v6R/0iROhDURE63vevn4XYpIu
aT8Rnu94393zUrsqdqNzYPjTT6jGTxCKELOBoHey0mnBNrDD9UG+LFVaaydbq8g6
L3a9RkR1sPL81A8Sq3KIzJKPCsMYjh9XCEhGOO0CgYEArH/I8K23QNK50jcO8vVR
ubIWBqYnjniGV93de7KjYb4OFeHgtZpSPZrGFCQvLd/Z6dl9gVJvpZTaY4kDU+uA
fXxTEkJzGFEfYWoJRihtaHBz5pOb6E33xIyZBiflRtakYFfB4UqdJV5yy/Vc5d8X
QHeFSYl/FXpaRkdrACgq+gkCgYEAoGFPsjM8nscwz/ESU/5uXhaFrIlTbeuq3u8Z
Sqgu/gBDxHI9C5FNohj8Mb2ZkyiPYbql2roVAfRiuWuCVpX+N1OFisw3DXESze2t
m0Ai6P3bG1tqlk3tc99rYCVyruj3vthNAPtRM/5QBs4lmWg0bgoVfMENmVQsRLLT
LriGsU0CgYEAt3yTpLTzpO+RmMf0xmBz1m40mlNXF0a+Jm0cCQHPPL4DgChnR95g
NJskS/uHzKBatAhntmuWc/dqOOivihCkBPSN1IdRLyYwPe58VQHH2PPA213ydCW5
a1NZfl/w08dPgU3HCwStAnwpHJvh+AQOAevpvQBZQv9CLvy32FhaChE=
-----END RSA PRIVATE KEY-----

```

No processo de criação também é gerado um arquivo de configuração ( `~/.kube/config` ) contendo informações dos certificado, users e password do cluster.

----------

## Configuração do Cluster do Kubernetes

Para acessar o kubernetes é necessário configurar as credenciais de acesso. Conforme comandos abaixo:

```
kubectl config set-cluster dev-aws-cluster --server=https://k8s-dev.a.luizalabs.com --insecure-skip-tls-verify=true
kubectl config set-credentials dev-aws-cluster-admin --username=admin --password=VOpgP0Ggnty5mLcq
kubectl config set-context dev --cluster=dev-aws-cluster --user=dev-aws-cluster-admin
kubectl config use-context dev
```

Feito isso execute `kubectl cluster-info`, será mostrado o status de execução dos serviços do cluster kubernetes.
```
Kubernetes master is running at https://k8s-dev.a.luizalabs.com
Elasticsearch is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kube-dns
KubeRegistry is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kube-registry
kubernetes-dashboard is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
Grafana is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
InfluxDB is running at https://k8s-dev.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb

```


Após a etapa de instalação e configuração, é necessário seguir alguns passos para criação do kube-registry:

#### 1º Passo - Criar uma o persitent Volume, que será usado para armazenar as nossas images Docker.

É possível criar via console AWS ou através da cli `aws ec2`, conforme exemplo abaixo:

#####Command:

```aws ec2 create-volume --size 100 --region us-east-1 --availability-zone us-east-1a --volume-type gp2```

#####Output:

```
{
    "AvailabilityZone": "us-east-1a",
    "Attachments": [],
    "Tags": [],
    "VolumeType": "gp2",
    "VolumeId": "vol-890ddf22",
    "State": "creating",
    "SnapshotId": null,
    "CreateTime": "YYYY-MM-DDTHH:MM:SS.000Z",
    "Size": 100
}
```
```
aws ec2 create-tags --resources vol-890ddf22 --tags Key=Name,Value=kube-system-kube-registry-pv Key=Project,Value=PaaS Key=Env,Value=Dev Key=Area,Value=Arquitetura Key=Role,Value=PersistentVolume Key=Team,Value=Arquitetura

```
Depois de criar o volume, é necessário adicionar o volumeID e o tamanho do disco no `storage:` no arquivo resources/kube-system-kube-registry-pv.yaml

```yaml
kind: PersistentVolume
apiVersion: v1
metadata:
  name: kube-system-kube-registry-pv
  labels:
    kubernetes.io/cluster-service: "true"
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteOnce
  awsElasticBlockStore:
    fsType: "ext4"
    volumeID: vol-890ddf22
```
Altere também o tamanho do disco no arquivo resources/kube-system-kube-registry-pvc.yaml.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: kube-registry-pvc
  namespace: kube-system
  labels:
    kubernetes.io/cluster-service: "true"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
```

Feito as alterações, prossiga na execução dos arquivos yaml abaixo:

1º - resources/kube-system-kube-registry-pv.yml - persistent volume pra ser usado no docker registry
```bash
$ kubectl create -f resources/kube-system-kube-registry-pv.yml
```
2º - resources/kube-system-kube-registry-pvc.yml - persistent volume claim
```bash
$ kubectl create -f resources/kube-system-kube-registry-pvc.yml
```
3º - resources/kube-system-kube-registry-rc.yml - docker registry replication controller
```bash
$ kubectl create -f resources/kube-system-kube-registry-rc.yml
```
4º - resources/kube-system-kube-registry-svc.yml - service
```bash
$ kubectl create -f resources/kube-system-kube-registry-svc.yml
```

Próximo passo é logar no master com o chave gerada para executar os comandos abaixo:

```bash
mkdir -p /etc/kubernetes/addons/registry
cat > /etc/kubernetes/addons/registry/kube-system-kube-registry-svc.yaml << EOF_KUBE_REGISTRY
apiVersion: v1
kind: Service
metadata:
  name: kube-registry
  namespace: kube-system
  labels:
    k8s-app: kube-registry
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeRegistry"
spec:
  selector:
    k8s-app: kube-registry
  ports:
  - name: registry
    port: 5000
    protocol: TCP
EOF_KUBE_REGISTRY
```
**Este processo irá garantir que o serviço kube-registry estará sempre rodando.**

## Ambiente 

![Kubernetes-Topology](https://github.com/luizalabs/paas/blob/master/k8s-dev/topology/kubernetes-aws-dev.jpg)




