## Instalação do cluster do Kubernetes produção.##

**k8s-lab-env.sh** - Variáveis de ambiente para configurar o cluster kubernetes na AWS na Região SA-EAST-1a com 5 nodes. É possível alterar o número de nodes depois do cluster criado, para isso, edite o ASG(AutoScaling Group) gerado automaticamente com o nome de **kubernetes-minion-group-sa-east-1a**.

```
export KUBE_AWS_ZONE=sa-east-1a
export AWS_S3_REGION=sa-east-1
export NUM_NODES=2
export MASTER_SIZE=m3.medium
export NODE_SIZE=m3.large
export KUBE_ENABLE_NODE_PUBLIC_IP=true
export KUBE_ENABLE_CLUSTER_REGISTRY=true
export KUBERNETES_PROVIDER=aws
export KUBERNETES_SKIP_DOWNLOAD=true
export KUBE_ENABLE_NODE_AUTOSCALER=true
export KUBE_ENABLE_DAEMONSET=true
export NON_MASQUERADE_CIDR="172.16.0.0/14"
export SERVICE_CLUSTER_IP_RANGE="172.16.0.0/16"
export DNS_SERVER_IP="172.16.0.10"
export MASTER_IP_RANGE="172.19.0.0/24"
export CLUSTER_IP_RANGE="172.18.0.0/16"

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
Kubernetes master is running at https://52.67.56.57
Elasticsearch is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/elasticsearch-logging
Heapster is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/heapster
Kibana is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/kibana-logging
KubeDNS is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/kube-dns
kubernetes-dashboard is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
Grafana is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/monitoring-grafana
InfluxDB is running at https://52.67.56.57/api/v1/proxy/namespaces/kube-system/services/monitoring-influxdb


Installation successful!
```

Foi criado uma chave ssh para conectar no master e nos nodes. Ela fica armazenada no diretório .ssh do HOME  do user que executou o `k8s-lab-install.sh`.
O nome gerado é `~/.ssh/kube_aws_rsa `
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA31jnMvugeHlk+qAhheQd/wcSpjE2JZgvH4j5Kjv1f/ETgh8N
zPkdjA3l4lutV4akTQNOWxYeACpWxO01oeroN23a4qRKoly0LxPDbnUkze8lYhEm
cgrouCZMvM2iVi9aUzc+C9oOf09MC5GDdsUtw5mHwazTeS6/1izsnLIE5ghpSOAZ
w465CjjqVKDceb04M6piVxty4Dgpd8CcLPWA/l6GA7WUSiG8P7KnnCcbsONbc8Z4
bOac3+EObtjk8VhBYVTiKXrLy4le4PdFe+7o43LhQ+CRjJY/TXacSCyzQLiO3AtK
kEDf6aiAoag69rC8Cw+RmHduJ9ZVsbof9D9g3QIDAQABAoIBAQC5FS46dITNcSO6
ARrmnlwxK4ZLzHone06JnnOEiT1vMbLN4LgKBOJ0XkwDYZ9q6bubykKTfueLkYpo
FH6eWFDpJhL+P9FNbO8wF/uavF6sfSIIe1fVp66kc7ChwPJm40TBswSqw5/B5k+J
QqOjt9Ctx7vVRkfUo2X7sSy+yVL/36oPzejBs1GYcDlkrO2lvD5YzPOD1cmhRckn
hsHhjJCmqeEt9mjWOzJwWBz3r/Df5h5bD+NPyXVxeQIOhRgesYUYNJfbOoULvYnn
jX19Spk9A/QZxINC0tCrirs9IyUPjHjbVhKx5QaDq1LFNbtuWM7UJgjEQoPi0HJ2
tF0oyWDhAoGBAPYL6OqXSbv+7AgMXhXQjk4kPBNefm5RG9vedY7K+XZAzRTsL9az
QviC526e5+5RJVoR5KbCaIMBcYJOP/gKx+tNx2p1P6Hjj7s11/FwNPMkSfm6fuR3
T9D7bxIZW8GJWmtGSGikPcr/dH6qwJ2aa2s+139DahTf0enBlcZaepyZAoGBAOhh
6ruJYB6X2gFnLUs6TUQPXkLmcySqCdGETtJFSiv1FGVLSBJBSDoutqyARWbHbSAY
rkzsURQqYb9n6nwVE4V5f6zA7Gs+EcGzDcTyP4IGihoOLzEHgEpEkdtdLweysaHO
dD4ZjGCvdgDIlIbcern7HyVGheJB05cky6uoPyzlAoGAK1YHrSH+a61Ht1cuTxtP
pTW+mw6+qAGDLaOuAvt/GrPpKzE6b89dEVZFGvfEE/aY5JdYNYLyU3iynGfZ3n9P
nQMzB24lSgdBrpCaOz0vJlAS83hEl0rodP+EHaT3w0vE8DYN1NhwxcteolzC1j93
ggksfY7LJWgSmeBt1+ThyakCgYBAQ7PyoQ8f5sa3VJ1GDJ2JkOZ0gd6d6RhGBNz3
cYoNlf1J9gO8aKlczcZe9io7PjODiM/LHh9eAEp/8HD8V6fKZctVLxUyozut4aKR
MJ03jC9s/Tc3y+dCoqvjimWTL2vIX5vuAIq9wkhT9yR8F0UpYbU/Tb2A0Mt/JQHe
Ou0dsQKBgQC3jQ4GEDtuysX8PWTp0VnAX2vG7xyT2bwYuxDFO/f+krMD6GdieGfZ
JOt1FocOs8PR3MLLdyEd6dZKIZ2ai6ZO0Eedr8/MvTJqOlJMrxnSphfWzG+8LYw7
XCz9+KbG29600kd/aITafCdpeEsZaSUl8VTI4OInO/59mx3449q1/Q==
-----END RSA PRIVATE KEY-----

```

No processo de criação também é gerado um arquivo de configuração ( `~/.kube/config` ) contendo informações dos certificado, users e password do cluster.

----------

## Configuração do Cluster do Kubernetes

Para acessar o kubernetes é necessário configurar as credenciais de acesso. Conforme comandos abaixo:

```
kubectl config set-cluster luizalabs-aws-cluster --server=https://k8s.a.luizalabs.com --insecure-skip-tls-verify=true
kubectl config set-credentials luizalabs-aws-cluster-admin --username=admin --password=0N40KAN6Z4xUsAl5
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
    "AvailabilityZone": "sa-east-1a",
    "Encrypted": false,
    "VolumeType": "gp2",
    "VolumeId": "vol-9107e531",
    "State": "creating",
    "Iops": 300,
    "SnapshotId": "",
    "CreateTime": "2016-05-25T19:52:10.719Z",
    "Size": 100
}

```
```
aws ec2 create-tags --resources vol-9107e531 --tags Key=Name,Value=kube-system-kube-registry-pv Key=Project,Value=PaaS Key=Env,Value=Prod Key=Area,Value=Arquitetura Key=Role,Value=Prod Key=Env,Value=PersistentVolume Key=Team,Value=Arquitetura

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
    volumeID: vol-c2515863
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

![Kubernetes-Topology](https://github.com/luizalabs/paas/blob/master/k8s/topology/topology.jpg)




