## Instalação do cluster do Kubernetes

k8s-lab-env.sh - variáveis de ambiente que configuram o cluster, para instalação do cluster na AWS
k8s-lab-install.sh - script que lê o k8s-lab-env.sh e dispara o instalador do kubernetes

O `k8s-lab-env.sh` já habilita o docker registry no cluster, mas para subir o POD do mesmo, habilitar um serviço e atrelar um persistent volume ao POD é necessário rodar um kubectl create nos seguintes arquivos:

## Configuração do Cluster do Kubernetes

###Master
Para cada um dos arquivos a chamada para o create será a mesma, exemplo:

1º - resources/kube-system-kube-registry-pv.yml - persistent volume pra ser usado no docker registry

    `kubectl create -f resources/kube-system-kube-registry-pv.yml`
    
2º - resources/kube-system-kube-registry-pvc.yml - persistent volume claim

    `kubectl create -f resources/kube-system-kube-registry-pvc.yml`
    
3º - resources/kube-system-kube-registry-rc.yml - docker registry replication controller

    `kubectl create -f resources/kube-system-kube-registry-rc.yml`
    
4º - resources/kube-system-kube-registry-svc.yml - service

    `kubectl create -f resources/kube-system-kube-registry-svc.yml`
    

####Nodes (Minions)
resources/kube-registry-proxy.yaml - Este arquivo precisa estar no diretório /etc/kubernetes/manifests para criação automática do pod que irá expor a porta :5000 do registry.


