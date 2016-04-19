## Instalação do cluster do Kubernetes

k8s-lab-env.sh - variáveis de ambiente que configuram o cluster, para instalação do cluster na AWS
k8s-lab-install.sh - script que lê o k8s-lab-env.sh e dispara o instalador do kubernetes

O `k8s-lab-env.sh` já habilita o docker registry no cluster, mas para subir o POD do mesmo, habilitar um serviço e atrelar um persistent volume ao POD é necessário rodar um kubectl create nos seguintes arquivos:

###Master

resources/kube-system-kube-registry-rc.yml - docker registry replication controller
resources/kube-system-kube-registry-pv.yml - persistent volume pra ser usado no docker registry
resources/kube-system-kube-registry-pvc.yml - persistent volume claim
resources/kube-system-kube-registry-svc.yml - service

###Nodes (Minions)
resources/kube-registry-proxy.yaml - 

Para cada um dos arquivos a chamada para o create será a mesma, exemplo:

    kubectl create -f resources/kube-system-kube-registry-rc.yml
