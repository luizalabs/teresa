Instalaç WeaveScope no Kubernetes
=
---
**Antes de começ**

        O cluster deve suportar DaemonSets. DaemonSets sãnecessáas para garantir que cada nóbernetes pode executar um **Scope probe**.

        Para habilitar DaemonSets em um cluster existente, adicione o `--runtime-config=extensions/v1beta1/daemonsets=true` argumento para a configuraç do apiserver. Isto éormalmente encontrado no /etc/kubernetes/manifest/kube-apiserver.manifestfile apósso, reinicie o serviç
> **Note:** Se vocêstiver criando um novo cluster, defina KUBE_ENABLE_DAEMONSETS = true na configuraç do seu cluster.

**Instalaç**

        Ãecomendado executar a instalaç do weavescope, atravé usando a cli kubectl, conforme mostrado abaixo.

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
weavescope-app-s5iqf                   1/1       Running   0          1d
weavescope-probe-ea7xa                 1/1       Running   0          1d
weavescope-probe-hkks7                 1/1       Running   0          1d
```
> **Note:** Apólguns segundos o status deve estar em "**Running**".

Para termos acesso externo com autenticaç e SSL, serásado um serviçcom NGINX fazendo Reverse Proxy para o WeaveScope.


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

