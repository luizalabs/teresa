## Services

**Get**
```bash
$ kubectl get services
NAME         CLUSTER-IP     EXTERNAL-IP       PORT(S)    AGE
hello-node   10.3.241.222   104.197.137.216   8080/TCP   11m
```

## Deployment

**Get**
```bash
$ kubectl get deployment
NAME         DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
hello-node   4         4         4            4           25m
```

**Describe**
```bash
$ kubectl describe deployments
```

**Aplicando alterações (utilizado para rollback, scale, etc)**
```bash
$ kubectl apply -f nginx-deployment.yaml
```

**Expor o deployment publicamente**
```bash
$ kubectl expose deployment hello-node --type="LoadBalancer"
```

**Scale (tem o mesmo efeito alterando o arquivo deployment)**
```bash
$ kubectl scale deployment hello-node --replicas=4
```

**Editando o deployment**
```bash
# Quando o arquivo é alterado, automaticamente o scheduller nota as alterações e já começo a aplicar as alterações
$ kubectl edit deployment hello-node
```

## Replicaset

**Get**
```bash
$ kubectl get replicasets
```

## Pods

**Get**
```bash
$ kubectl get pods
NAME                          READY     STATUS    RESTARTS   AGE
hello-node-2628160756-5bod2   1/1       Running   0          1m
hello-node-2628160756-g8wmu   1/1       Running   0          1m
hello-node-2628160756-uzz03   1/1       Running   0          1m
hello-node-2628160756-wllk4   1/1       Running   0          25m
```

**Get dos pods + labels**
```bash
$ kubectl get pods --show-labels
NAME                          READY     STATUS    RESTARTS   AGE       LABELS
hello-node-2628160756-0w05s   1/1       Running   0          34m       pod-template-hash=2628160756,run=hello-node
hello-node-2628160756-i5o0o   1/1       Running   0          34m       pod-template-hash=2628160756,run=hello-node
```

## Build

**Build docker image (Google Cloud)**
```bash
$ docker build -t gcr.io/PROJECT_ID/hello-node:v2 . 
$ gcloud docker push gcr.io/PROJECT_ID/hello-node:v2
```

## Run

**Cria deployment e executa o container**
```bash
$ kubectl run hello-node --image=gcr.io/PROJECT_ID/hello-node:v1 --port=8080
deployment "hello-node" created
```

## Delete (destroy)

**Destruíndo serviço e deployment**
```bash
$ kubectl delete service,deployment hello-node
```
