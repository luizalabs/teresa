#!/bin/bash

source settings.sh
source kubernetes-helpers.sh
export AWS_DEFAULT_OUTPUT=text
if [ $# != 1 ]; then
    echo "You need to specify kubernetes cluster_ID. EX:kubernetes"
    exit
fi
vpc_id=$(get_vpc_id $1)
echo $vpc_id
instance_id=$(get_kube_master $vpc_id)
echo "Kubernetes MasterID = " $instance_id
