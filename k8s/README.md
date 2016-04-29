## Instalação do cluster do Kubernetes produção.##

**k8s-lab-env.sh** - Variáveis de ambiente para configurar o cluster kubernetes na AWS na Região SA-EAST-1a com 5 nodes. É possível alterar o número de nodes depois do cluster criado, para isso, edite o ASG(AutoScaling Group) criado automaticamente com o nome de **kubernetes-minion-group-sa-east-1a**.

```
export KUBE_AWS_ZONE=sa-east-1a
export AWS_S3_REGION=sa-east-1
export NUM_NODES=5
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

Este processo irá criar passar por 11 etapas, descritas abaixo:

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

* Mandatoriamente o Kubernetes usa um VPC com bloco 172.20.0.0/16 e um SUBNET 172.20.0.0/24, caso você já tenha criado estes recursos, você precisará exportar as variáveis VPC_ID e SUBNET_ID com os ids dos mesmos.

Ao final da execução, será exibido o endereço do Master e seus serviços. 

```
Kubernetes master is running at https://52.67.9.216
Elasticsearch is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/kube-dns
kubernetes-dashboard is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
Grafana is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
InfluxDB is running at https://52.67.9.216/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb

Installation successful!
```

Foi criado uma chave ssh para conectar no master e nos nodes. Ela fica armazenada no diretório .ssh do HOME  do user que executou o `k8s-lab-install.sh`.
O nome gerado é `~/.ssh/kube_aws_rsa `

No processo de criação também é gerado um arquivo de configuração ( `~/.kube/config` ) contendo informações dos certificado, users e password do cluster.


----------

## Configuração do Cluster do Kubernetes

Para acessar o kubernetes é necessário configurar as credenciais de acesso. Conforme comandos abaixo:

```
kubectl config set-cluster luizalabs-aws-cluster --server=https://k8s.a.luizalabs.com --insecure-skip-tls-verify=true
kubectl config set-credentials luizalabs-aws-cluster-admin --username=admin --password=AUHK0P1NMAtMKEK6
kubectl config set-context luizalabs-aws-cluster-context --cluster=luizalabs-aws-cluster --user=luizalabs-aws-cluster-admin
kubectl config use-context luizalabs-aws-cluster-context
```

Feito isso execute `kubectl cluster-info`, será mostrado o status de execução dos serviços do cluster kubernetes.
```
Kubernetes master is running at https://k8s.a.luizalabs.com
Elasticsearch is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kube-dns
KubeRegistry is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kube-registry
kubernetes-dashboard is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
Grafana is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
InfluxDB is running at https://k8s.a.luizalabs.com/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb
```



Após a etapa de instalação e configuração, é necessário seguir alguns passos para criação do kube-registry:

#### 1º Passo - Criar uma o persitent Volume, que será usado para armazenar as nossas images Docker.

É possível criar via console AWS ou através da cli `aws ec2`, conforme exemplo abaixo:

#####Command:

```aws ec2 create-volume --size 100 --region sa-east-1 --availability-zone sa-east-1a --volume-type gp2```

#####Output:

```
{
    "AvailabilityZone": "us-east-1a",
    "Attachments": [],
    "Tags": [],
    "VolumeType": "gp2",
    "VolumeId": "vol-c2515863",
    "State": "creating",
    "SnapshotId": null,
    "CreateTime": "YYYY-MM-DDTHH:MM:SS.000Z",
    "Size": 100
}
```
```
aws ec2 create-tags --resources vol-c2515863 --tags Key=Name,Value=kube-system-kube-registry-pv Key=Project,Value=PaaS Key=Env,Value=Prod Key=Area,Value=Arquitetura Key=Role,Value=Prod Key=Env,Value=PersistentVolume Key=Team,Value=Arquitetura

```
Depois de criar o volume, é necessário adicionar o volumeID e o tamanho do disco no `storage:` no arquivo resources/kube-system-kube-registry-pv.yaml

```
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
    volumeID: vol-c2515863
```
Altere também o tamanho do disco no arquivo resources/kube-system-kube-registry-pvc.yaml.

```
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
    

####Nodes (Minions)
resources/kube-registry-proxy.yaml - Este arquivo precisa estar no diretório /etc/kubernetes/manifests para criação automática do pod que irá expor a porta :5000 do registry.
