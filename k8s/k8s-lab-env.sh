export KUBERNETES_PROVIDER=aws
export KUBE_ENABLE_NODE_AUTOSCALER=true
export KUBE_AWS_ZONE=sa-east-1a
export NUM_NODES=5
export MASTER_SIZE=m3.medium
export NODE_SIZE=m3.large

# https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/registry#how-it-works
export KUBE_ENABLE_CLUSTER_REGISTRY=true
export ENABLE_NODE_PUBLIC_IP=true

