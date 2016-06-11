#!/bin/bash
# Name: kubernetes-helpers.sh
# Description: Global function for kubernetes
# Autor: Eraldo S.Bernardino
# Date: 01/06/2016

#Functions

function get_vpc_id {
    aws ec2 describe-vpcs --filters \
    Name=tag:Name,Values=kubernetes-vpc \
    Name=tag:KubernetesCluster,Values=${CLUSTER_ID} \
    --query Vpcs[].VpcId 2>> $LOG
}
function get_kube_master {
    local vpc_id=$1
    aws ec2 describe-instances --filters \
    Name=instance-state-name,Values=running \
    Name=vpc-id,Values=$vpc_id \
    Name=tag:KubernetesCluster,Values=kubernetes  \
    Name=tag:Role,Values=kubernetes-master \
    --query 'Reservations[].Instances[].InstanceId'
 }
function get_master_ebs_id {
    aws ec2 describe-volumes --filters \
    Name=availability-zone,Values=$AZ \
    Name=tag:Name,Values=kubernetes-master-pd \
    Name=tag:KubernetesCluster,Values=kubernetes \
    --query 'Volumes[].VolumeId' 2>> $LOG
}
function get_master_eip {
    local master_ebs_id=$1
    aws ec2 describe-tags --filters \
    Name=resource-id,Values=$master_ebs_id \
    Name=key,Values=kubernetes.io/master-ip \
    --query 'Tags[].Value' 2>> $LOG
}
function get_master_id {
    local vpc_id=$1
    aws ec2 describe-instances --filters \
    Name=vpc-id,Values=$vpc_id \
    Name=tag:KubernetesCluster,Values=$CLUSTER_ID  \
    Name=tag:Role,Values=$MASTER_NAME \
    --query 'Reservations[].Instances[].InstanceId' 2>> $LOG
}
function get_master_state {
    local instance_id=$1
    aws ec2 describe-instance-status --instance-ids $instance_id \
    --query 'InstanceStatuses[].InstanceState[].Code' 2>> $LOG
}
function get_master_pip {
    aws ec2 describe-addresses --filters \
    Name=private-ip-address,Values=$MASTER_PIP \
    --query 'Addresses[].InstanceId' 2>> $LOG
}
function get_eip_state {
    local eip=$1
    aws ec2 describe-addresses --public-ips $eip \
    --query 'Addresses[].InstanceId' 2>> $LOG
}
function get_master_sg_name {
    local vpc_id=$1
    aws ec2 describe-security-groups --filters \
    Name=vpc-id,Values=$vpc_id \
    Name=group-name,Values=kubernetes-master-kubernetes \
    Name=tag:KubernetesCluster,Values=kubernetes \
    --query 'SecurityGroups[].GroupId' 2>> $LOG
}
function get_master_subnet_id {
     local vpc_id=$1
     aws ec2 describe-subnets --filters \
     Name=tag:KubernetesCluster,Values=kubernetes \
     Name=availabilityZone,Values=$AZ \
     Name=vpc-id,Values=$vpc_id --query \
     'Subnets[].SubnetId' 2>> $LOG
}
function create_kube_master {
    local vpc_id=$1
    local master_subnet_id=$(get_master_subnet_id $vpc_id)
    local master_sg_name=$(get_master_sg_name $vpc_id)
    aws ec2 run-instances --image-id $MASTER_AMI_ID \
    --iam-instance-profile Name=$MASTER_NAME \
    --instance-type $MASTER_INSTANCE_TYPE \
    --subnet-id $master_subnet_id \
    --private-ip-address $MASTER_PIP \
    --key-name $MASTER_KEY_NAME \
    --security-group-ids $master_sg_name \
    --associate-public-ip-address \
    --block-device-mappings '[{"DeviceName":"/dev/xvda","Ebs":{"DeleteOnTermination":true,"VolumeSize":8,"VolumeType":"gp2"}} ,{"DeviceName": "/dev/sdc","VirtualName":"ephemeral0"},{"DeviceName": "/dev/sdd","VirtualName":"ephemeral1"},{"DeviceName": "/dev/sde","VirtualName":"ephemeral2"},{"DeviceName": "/dev/sdf","VirtualName":"ephemeral3"}]' \
    --user-data fileb://$MASTER_USER_DATA \
    --query 'Instances[].InstanceId' 2>> $LOG
}
function wait_for_instance_state {
    local instance_id=$1
    local state=$2
    while true; do
       instance_state=$(aws ec2 describe-instances --instance-ids ${instance_id} --query Reservations[].Instances[].State.Name 2>> $LOG )
        if [[ "$instance_state" == "${state}" ]]; then
            echo -e "$INFO New kubernetes master $GREEN$instance_id$NONE is $GREEN running$NONE."
            break
        else
            echo -e "$INFO Waiting for instance $RED${instance_id}$NONE to be ${state} (currently $RED${instance_state})"$NONE
            echo -e "$INFO Sleeping for 3 seconds..."
            sleep 3
        fi
    done
}
function get_eip_allocationid {
    local eip=$1
    eip_allocationid=$(aws ec2 describe-addresses --public-ips $eip \
    --query 'Addresses[].AllocationId'  2>> $LOG)
    echo -e "$INFO ElasticIP AllocationId is $GREEN$eip_allocationid$NONE."
}
function exec_associate_eip {
    local instance_id=$1
    local eip_allocationid=$2
    local master_eip=$3
    aws ec2 associate-address --instance-id $instance_id \
    --allocation-id $eip_allocationid  >> $LOG  2>&1
    echo -e "$INFO Associating EIP $GREEN$master_eip$NONE in new Kubernetes-Master $GREEN$instance_id$NONE."
}
function exec_attach_volume {
    local instance_id=$1
    local master_ebs_id=$2
    aws ec2 attach-volume --volume-id $master_ebs_id \
    --device /dev/sdb --instance-id $instance_id >> $LOG 2>&1
    echo -e "$INFO Associating EBS Volume ID $GREEN$MASTER_EBS_ID$NONE in new Kubernetes-Master $GREEN$instance_id$NONE."
}
function create_master_tags {
    local instance_id=$1
    local environment=$2
    aws ec2 create-tags --resources $instance_id --tags \
    Key=Name,Value=kubernetes-master \
    Key=Role,Value=kubernetes-master \
    Key=KubernetesCluster,Value=kubernetes \
    Key=Project,Value=PaaS \
    Key=Area,Value=Coreservices \
    Key=Env,Value=$environment \
    Key=Team,Value=Arquitetura 2>> $LOG
    echo -e "$INFO Creating kubernetes-master tags."
}
