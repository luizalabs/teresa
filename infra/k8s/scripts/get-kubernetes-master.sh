#!/bin/bash


export AWS_DEFAULT_OUTPUT=text
if [ $# != 1 ]; then

    echo "You need to specify kubernetes cluster_ID. EX:kubernetes"
    exit

fi




function get_kube_master {
	local vpc_id=$1
	aws ec2 describe-instances --filters \
	Name=instance-state-name,Values=running \
	Name=vpc-id,Values=$vpc_id \
	Name=tag:KubernetesCluster,Values=kubernetes  \
	Name=tag:Role,Values=kubernetes-master \
	--query 'Reservations[].Instances[].InstanceId'
} 

function get_VPC_ID {
    local cluster_id=$1
    aws ec2 describe-vpcs --filters \
    Name=tag:Name,Values=kubernetes-vpc \
    Name=tag:KubernetesCluster,Values=$cluster_id \
    --query Vpcs[].VpcId
}

vpc_id=$(get_VPC_ID $1)

instance_id=$(get_kube_master $vpc_id)

echo "Kubernetes MasterID = " $instance_id
