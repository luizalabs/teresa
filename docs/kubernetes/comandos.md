## Resources

É possível utilizar os comandos descritos nesta páginas com toda (ou quase toda) a lista de resources abaixo (não está completa, apenas com os mais utilizados).  

`TODO: colocar um link aqui para uma descrição completa de todos os resources`

- deployments
- events (alias 'ev')
- endpoints (alias 'ep')
- nodes (alias 'no')
- namespaces (alias 'ns')
- pods (alias 'po')
- persistentvolumes (alias 'pv')
- persistentvolumeclaims (alias 'pvc')
- replicasets (alias 'rs')
- replicationcontrollers (alias 'rc')
- services (alias 'svc')

## Namespace

Todo o resource dentro do kubernetes fica registrado em um namespace específico.  
Ou seja, para se obter por exemplo uma lista de serviços registrados, é necessário especificar a qual namespace se deseja esta lista. Se não for especificado o namespace, o namespace utilizado é o "default".  
O parâmetro `--namespace` é global, ou seja, serve para todos os comandos do kubectl ligados a resources.  
O mesmo pode ser substituído por `--all-namespaces` que lista todos os resources registrados.  

**Exemplos:**  

Retorna informações do namespace "kube-system".  
```bash
$ kubectl get services --namespace kube-system
NAME                    CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
elasticsearch-logging   10.0.23.124    <none>        9200/TCP            20h
heapster                10.0.114.110   <none>        80/TCP              20h
kibana-logging          10.0.135.164   <none>        5601/TCP            20h
kube-dns                10.0.0.10      <none>        53/UDP,53/TCP       20h
kubernetes-dashboard    10.0.42.57     <none>        80/TCP              20h
monitoring-grafana      10.0.85.71     <none>        80/TCP              20h
monitoring-influxdb     10.0.143.80    <none>        8083/TCP,8086/TCP   20h
```
Retorna informações referentes a todos os namespaces, inclusive o namespace.  
```bash
$ kubectl get services --all-namespaces
NAMESPACE     NAME                    CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
default       kubernetes              10.0.0.1       <none>        443/TCP             21h
kube-system   elasticsearch-logging   10.0.23.124    <none>        9200/TCP            21h
kube-system   heapster                10.0.114.110   <none>        80/TCP              21h
kube-system   kibana-logging          10.0.135.164   <none>        5601/TCP            21h
kube-system   kube-dns                10.0.0.10      <none>        53/UDP,53/TCP       21h
kube-system   kubernetes-dashboard    10.0.42.57     <none>        80/TCP              21h
kube-system   monitoring-grafana      10.0.85.71     <none>        80/TCP              21h
kube-system   monitoring-influxdb     10.0.143.80    <none>        8083/TCP,8086/TCP   21h
```

## Comandos

### Get

Retorna uma lista resumida do resource.  

```bash
$ kubectl get <resource>

# Namespace
$ kubectl get services --namespace kube-system
NAME                    CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
elasticsearch-logging   10.0.23.124    <none>        9200/TCP            20h
heapster                10.0.114.110   <none>        80/TCP              20h
...

# Namespace + Labels
$ kubectl get services --namespace kube-system --show-labels
NAME                    CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE       LABELS
heapster                10.0.114.110   <none>        80/TCP              21h       kubernetes.io/cluster-service=true,kubernetes.io/name=Heapster
kubernetes-dashboard    10.0.42.57     <none>        80/TCP              21h       k8s-app=kubernetes-dashboard,kubernetes.io/cluster-service=true
...
```

### Describe

Retornar informações detalhadas do serviço.  
Se o serviço não for informado serão exibidos todos os serviços.  

```bash
$ kubectl describe <resource> <resource-name>

$ kubectl describe service kubernetes-dashboard  --namespace kube-system
Name:     kubernetes-dashboard
Namespace:    kube-system
Labels:     k8s-app=kubernetes-dashboard,kubernetes.io/cluster-service=true
Selector:   k8s-app=kubernetes-dashboard
Type:     ClusterIP
IP:     10.0.42.57
Port:     <unset> 80/TCP
Endpoints:    10.244.0.4:9090
Session Affinity: None
No events.
```


### Create

Cria um novo resource.  
A flag `--record` registra o comando executado como anotações no resource. **Muito importante principalmente para deploys.**  

```bash
$ kubectl create -f <file-descrevendo-resource> --record

$ kubectl create -f nginx-deployment.yaml --record
```

Exemplo nginx-deployment.yaml
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
```

### Apply

Alterando um resource já criado.  
Segue o mesmo modelo do create, porem cria alterações em resources já existentes.  

```bash
$ kubectl apply -f nginx-deployment.yaml
```

### Delete

Remove um resource.  
`Atenção:` Alguns resources como por exemplo o deployment executam por default o "cascade delete".  

```bash
$ kubectl delete service <nome-do-resource>
```

### Expose

Criar um serviço expondo um deployment.  

```bash
$ kubectl expose deployment <nome-do-deployment> --type="LoadBalancer"
```

### Scale

Escala um deployment.  
Faz o mesmo efeito que um apply direto no deployment. Em nível de API url seria um PATCH.  

```bash
$ kubectl scale deployment <nome-do-deployment> --replicas=4
```

### Edit

Edita um resource específico.  
Quando o arquivo é alterado, automaticamente o scheduller nota as alterações e já começo a aplicalas.  
Tem o mesmo efeito que um apply a partir de um file local.  

```bash
$ kubectl edit <resource> <nome-do-resource>
```

### Run
[kubernetes.io/docs/run/](http://kubernetes.io/docs/user-guide/kubectl/kubectl_run/)
Cria um deployment a partir de uma imagem.  

```bash
$ kubectl run <app-name> --image=<imagem+versão> --port=<porta>
```

---
## Docker

### Docker build

Build de um docker gerando uma imagem com tag para ser futuramente enviada para um registry privado.  

```bash
# Dentro do diretório do app que contem o Dockerfile
$ docker build -t <url-docker-registry>:<versão> .
```

### Docker push

Envia a imagem local para o docker registry

```bash
$ docker push <url-docker-registry>:<versão>
```

Exemplo de envio para o docker registry do google compute engine

```bash
# O comando gcloud antes do docker push é um wrapper para gerenciar autenticação no gcloud
$ gcloud docker push gcr.io/meu-projeto/hello-app:v1
```

