Instalação WeaveScope no Kubernetes
=
---
**Antes de começar**

        O cluster deve suportar DaemonSets. DaemonSets são necessários para garantir que cada nó do kubernetes possa executar um **Scope probe**.

        Para habilitar DaemonSets em um cluster existente, adicione o `--runtime-config=extensions/v1beta1/daemonsets=true` argumento para a configuração do apiserver. Isto é normalmente encontrado no /etc/kubernetes/manifest/kube-apiserver.manifestfile após isso, reinicie o serviço
> **Note:** Se você estiver criando um novo cluster, defina KUBE_ENABLE_DAEMONSETS = true nas variáveis de ambiente do seu cluster.

**Instalação**

        Recomendado executar a instalação do weavescope, através usando a cli kubectl, conforme mostrado abaixo.

```
kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false --namespace=kube-infra
```
Segue abaixo o scope manifest `weavescope.yaml`,  executado no passo anterior.

```json
apiVersion: v1
kind: List
items:
  - metadata:
      labels:
        app: weavescope
        weavescope-component: weavescope-app
      name: weavescope-app
    apiVersion: v1
    kind: ReplicationController
    spec:
      replicas: 1
      template:
        metadata:
          labels:
            app: weavescope
            weavescope-component: weavescope-app
        spec:
          containers:
            - name: weavescope-app
              image: 'weaveworks/scope:0.14.0'
              args:
                - '--no-probe'
              ports:
                - containerPort: 4040
  - metadata:
      labels:
        app: weavescope
        weavescope-component: weavescope-app
      name: weavescope-app
    apiVersion: v1
    kind: Service
    spec:
      ports:
        - name: app
          port: 80
          targetPort: 4040
          protocol: TCP
      selector:
        app: weavescope
        weavescope-component: weavescope-app
  - metadata:
      labels:
        app: weavescope
        weavescope-component: weavescope-probe
      name: weavescope-probe
    apiVersion: extensions/v1beta1
    kind: DaemonSet
    spec:
      template:
        metadata:
          labels:
            app: weavescope
            weavescope-component: weavescope-probe
        spec:
          hostPID: true
          hostNetwork: true
          containers:
            - name: weavescope-probe
              image: 'weaveworks/scope:0.14.0'
              args:
                - '--no-app'
                - '--probe.docker.bridge=docker0'
                - '--probe.docker=true'
                - '--probe.kubernetes=true'
                - '$(WEAVESCOPE_APP_SERVICE_HOST):$(WEAVESCOPE_APP_SERVICE_PORT)'
              securityContext:
                privileged: true
              resources:
                limits:
                  cpu: 50m
              volumeMounts:
                - name: docker-sock
                  mountPath: /var/run/docker.sock
          volumes:
            - name: docker-sock
              hostPath:
                path: /var/run/docker.sock
```

**Verifique o status dos pods executando o comando abaixo:**

```
 **kubectl get pods | grep weave**
 kubectl get pods --namespace=kube-infra | grep weave
weavescope-app-idui1           1/1       Running   0          1d
weavescope-probe-04ky9         1/1       Running   0          1d
weavescope-probe-ewjmv         1/1       Running   0          1d
weavescope-probe-gjyql         1/1       Running   0          1d

```
> **Note:** Após alguns segundos o status deve estar em "**Running**".

Para termos acesso externo com autenticação e SSL, usaremos um serviço com NGINX fazendo Reverse Proxy para o WeaveScope.


```
kubectl describe svc weavescope-app
Name:                   weavescope-app
Namespace:              default
Labels:                 app=weavescope,weavescope-component=weavescope-app
Selector:               app=weavescope,weavescope-component=weavescope-app
Type:                   ClusterIP
IP:                     10.0.89.119
Port:                   app     80/TCP
Endpoints:              10.244.0.6:4040
Session Affinity:       None
```

