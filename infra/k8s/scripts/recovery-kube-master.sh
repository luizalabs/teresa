#!/bin/bash
# Name: recorevy-kube-master.sh
# Description: Recovery Kubernetes Master
# Autor: Eraldo S.Bernardino
# Date: 01/06/2016
export AWS_DEFAULT_OUTPUT=text

#Color
NONE='\033[00m'
RED='\033[01;31m'
GREEN='\033[01;32m'
YELLOW='\033[01;33m'
PURPLE='\033[01;35m'
CYAN='\033[01;36m'
WHITE='\033[01;37m'
BOLD='\033[1m'

#Notification
INFO="$GREEN INFO -$NONE"
WARN="$YELLOW WARN -$NONE"
ERRO="$RED ERRO -$NONE"


#Global Variables
ZONE="us-west-2"
AZ="us-west-2a"
CLUSTER_ID="kubernetes"
MASTER_NAME=$CLUSTER_ID"-master"
MASTER_PIP="172.20.0.9"
MASTER_USER_DATA="us-west-2-kubernentes/master-user-data.gz"
MASTER_INSTANCE_TYPE="m3.medium"
MASTER_KEY_NAME="kubernetes-3a0ab4c58a85f9173c85e7aaf2df8baa"
MASTER_AMI_ID="ami-7840ac18"
LOG="recovery-kube-master.log"

#Functions

function get_VPC_ID {
    aws ec2 describe-vpcs --filters \
    Name=tag:Name,Values=kubernetes-vpc \
    Name=tag:KubernetesCluster,Values=${CLUSTER_ID} \
    --query Vpcs[].VpcId 2>> $LOG
}
function get_MASTER_EBS_ID {
    aws ec2 describe-volumes --filters \
    Name=availability-zone,Values=$AZ \
    Name=tag:Name,Values=kubernetes-master-pd \
    Name=tag:KubernetesCluster,Values=kubernetes \
    --query 'Volumes[].VolumeId' 2>> $LOG
}
function get_MASTER_EIP {
    local MASTER_EBS_ID=$1
    aws ec2 describe-tags --filters \
    Name=resource-id,Values=$MASTER_EBS_ID \
    Name=key,Values=kubernetes.io/master-ip \
    --query 'Tags[].Value' 2>> $LOG
}
function get_MASTER_ID {
    local VPC_ID=$1
    aws ec2 describe-instances --filters \
    Name=vpc-id,Values=$VPC_ID \
    Name=tag:KubernetesCluster,Values=$CLUSTER_ID  \
    Name=tag:Role,Values=$MASTER_NAME \
    --query 'Reservations[].Instances[].InstanceId' 2>> $LOG
}
function get_MASTER_STATE {
    local instance_id=$1
    aws ec2 describe-instance-status --instance-ids $instance_id \
    --query 'InstanceStatuses[].InstanceState[].Code' 2>> $LOG
}
function get_MASTER_PIP {
    aws ec2 describe-addresses --filters \
    Name=private-ip-address,Values=$MASTER_PIP \
    --query 'Addresses[].InstanceId' 2>> $LOG
}
function get_EIP_STATE {
    EIP=$1
    aws ec2 describe-addresses --public-ips $EIP \
    --query 'Addresses[].InstanceId' 2>> $LOG
}
function get_MASTER_SG_NAME {
    local VPC_ID=$1
    aws ec2 describe-security-groups --filters \
    Name=vpc-id,Values=$VPC_ID \
    Name=group-name,Values=kubernetes-master-kubernetes \
    Name=tag:KubernetesCluster,Values=kubernetes \
    --query 'SecurityGroups[].GroupId' 2>> $LOG
}
function get_MASTER_SUBNET_ID {
     local VPC_ID=$1
     aws ec2 describe-subnets --filters \
     Name=tag:KubernetesCluster,Values=kubernetes \
     Name=availabilityZone,Values=$AZ \
     Name=vpc-id,Values=$VPC_ID --query \
     'Subnets[].SubnetId' 2>> $LOG
}
function create_KUBE_MASTER {
    local VPC_ID=$1
    MASTER_SUBNET_ID=$(get_MASTER_SUBNET_ID $VPC_ID)
    #echo -e "$INFO Will be use Master SubnetId $GREEN$MASTER_SUBNET_ID$NONE."
    MASTER_SG_NAME=$(get_MASTER_SG_NAME $VPC_ID)
    #echo -e "$INFO Will be use Master SecurityGroupsId $GREEN$MASTER_SG_NAME$NONE."
    aws ec2 run-instances --image-id $MASTER_AMI_ID \
    --iam-instance-profile Name=$MASTER_NAME \
    --instance-type $MASTER_INSTANCE_TYPE \
    --subnet-id $MASTER_SUBNET_ID \
    --private-ip-address $MASTER_PIP \
    --key-name $MASTER_KEY_NAME \
    --security-group-ids $MASTER_SG_NAME \
    --associate-public-ip-address \
    --block-device-mappings '[{"DeviceName":"/dev/xvda","Ebs":{"DeleteOnTermination":true,"VolumeSize":8,"VolumeType":"gp2"}} ,{"DeviceName": "/dev/sdc","VirtualName":"ephemeral0"},{"DeviceName": "/dev/sdd","VirtualName":"ephemeral1"},{"DeviceName": "/dev/sde","VirtualName":"ephemeral2"},{"DeviceName": "/dev/sdf","VirtualName":"ephemeral3"}]' \
    --user-data fileb://$MASTER_USER_DATA \
    --query 'Instances[].InstanceId' 2>> $LOG
}
function wait-for-instance-state {
    local instance_id=$1
    local state=$2
    while true; do
       instance_state=$(aws ec2 describe-instances --instance-ids ${instance_id} --query Reservations[].Instances[].State.Name 2>> $LOG )
        if [[ "$instance_state" == "${state}" ]]; then
            echo -e "$INFO New kubernetes master $GREEN$1$NONE is $GREEN running$NONE."
            break
        else
            echo -e "$INFO Waiting for instance $RED${instance_id}$NONE to be ${state} (currently $RED${instance_state})"$NONE
            echo -e "$INFO Sleeping for 3 seconds..."
            sleep 3
        fi
    done
}
function get_EIP_ALLocationId {
    local EIP=$1
    EIP_ALLocationId=$(aws ec2 describe-addresses --public-ips $EIP \
    --query 'Addresses[].AllocationId'  2>> $LOG)
    echo -e "$INFO ElasticIP AllocationId is $GREEN$EIP_ALLocationId$NONE."
}
function exec_associateEIP {
    local instance_id=$1
    local EIP_ALLocationId=$2
    local MASTER_EIP=$3
    aws ec2 associate-address --instance-id $instance_id \
    --allocation-id $EIP_ALLocationId  >> $LOG  2>&1
    echo -e "$INFO Associating EIP $GREEN$MASTER_EIP$NONE in new Kubernetes-Master $GREEN$instance_id$NONE."
}
function exec_ATTACH_VOLUME {
    local instance_id=$1
    local MASTER_EBS_ID=$2
    aws ec2 attach-volume --volume-id $MASTER_EBS_ID \
    --device /dev/sdb --instance-id $instance_id >> $LOG 2>&1
    echo -e "$INFO Associating EBS Volume ID $GREEN$MASTER_EBS_ID$NONE in new Kubernetes-Master $GREEN$instance_id$NONE."
}
function create_master_tags {
    local instance_id=$1
    aws ec2 create-tags --resources $instance_id --tags \
    Key=Name,Value=kubernetes-master \
    Key=Role,Value=kubernetes-master \
    Key=KubernetesCluster,Value=kubernetes \
    Key=Project,Value=PaaS Key=Env,Value=Prod \
    Key=Area,Value=Coreservices \
    Key=Team,Value=Arquitetura 2>> $LOG
    echo -e "$INFO Creating kubernetes-master tags."
}
banner() {
    msg="# $* #"
    edge=$(echo -e "$msg" | sed 's/./#/g')
    echo -e "$edge"
    echo -e "$msg"
    echo -e "$edge"
 }

clear
banner "                                      Kubernetes Recovery Master  ver.1.0                                               "
echo -e $CYAN"                                                Starting process..."$NONE

VPC_ID=$(get_VPC_ID)
MASTER_EBS_ID=$(get_MASTER_EBS_ID)
MASTER_ID=$(get_MASTER_ID $VPC_ID)
MASTER_EIP=$(get_MASTER_EIP $MASTER_EBS_ID)

# For tests
#MASTER_STATE="running"
#MASTER_STATE="stopped"
#MASTER_STATE=""
#MASTER_EIP="52.41.23.75"

echo  "----------------------------------------------------------------------------------------------------------------------------"
echo -e "$INFO Starting $(date +%T--%d-%m-%Y)."
echo -e "$INFO AWS Zone is $GREEN$ZONE$NONE."
echo -e "$INFO AWS Availability Zone is $GREEN$AZ$NONE."
echo -e "$INFO AWS VPC-ID is $GREEN$VPC_ID$NONE."
echo -e "$INFO Kubernetes ID is $GREEN$CLUSTER_ID$NONE."
echo -e "$INFO Kubernetes Master Name is $GREEN$MASTER_NAME$NONE."

if [[ -n "${MASTER_ID}" ]] ; then
    MASTER_STATE=$(get_MASTER_STATE $MASTER_ID)
    case "$MASTER_STATE" in
       0) echo -e "$INFO Kubernetes Master $GREEN$MASTER_ID$NONE with EIP $GREEN$MASTER_EIP$NONE is$RED Pending$NONE, wait and execute again.";;
      16) echo -e "$INFO Kubernetes Master $GREEN$MASTER_ID$NONE with EIP $GREEN$MASTER_EIP$NONE is$GREEN Running$NONE, Nothing to do.";;
      32) echo -e "$INFO Kubernetes Master $GREEN$MASTER_ID$NONE with EIP $GREEN$MASTER_EIP$NONE is$RED shutting-down$NONE, wait and execute again.";;
      64) echo -e "$WARN Kubernetes Master $GREEN$MASTER_ID$NONE with EIP $GREEN$MASTER_EIP$NONE is$RED Stopping$NONE."
          echo -e "$INFO To rebuild Kubernetes-Master is necessary that previous instance state be terminated or does not exist.";;
      80) echo -e "$WARN kubernetes Master $GREEN$MASTER_ID is $RED Stopped$NONE."
          echo -e "$INFO To rebuild Kubernetes-Master is necessary that previous instance state be terminated or does not exist.";;
    esac
else
    echo -e "$INFO Will be create new kubernetes Master."
    echo -e "$INFO Process for creating new Kubernetes Master starting."
    EIP_STATE=$(get_EIP_STATE $MASTER_EIP)
    if [[ -n "${EIP_STATE}" ]]; then
            echo -e "$ERRO Elastic IP $RED$MASTER_EIP$NONE allocated for InstanceID = $RED$EIP_STATE$NONE."

    else
        echo -e "$INFO ElasticIP $GREEN$MASTER_EIP$NONE is available."
        PIP_STATE=$(get_MASTER_PIP)

        if [[ -n "${PIP_STATE}" ]]; then
            echo -e "$ERRO Private IP $RED$MASTER_PIP$NONE allocated for InstanceID = $RED$PIP_STATE$NONE."
        else
            echo -e "$INFO Private IP $GREEN$MASTER_PIP$NONE is available."
            echo -e "$INFO Starting create new kubernetes master."
            MASTER_ID=$(create_KUBE_MASTER $VPC_ID)
            wait-for-instance-state $MASTER_ID "running"
            get_EIP_ALLocationId $MASTER_EIP
            exec_associateEIP $MASTER_ID $EIP_ALLocationId $MASTER_EIP
            exec_ATTACH_VOLUME $MASTER_ID $MASTER_EBS_ID
            create_master_tags $MASTER_ID
        fi
    fi
fi
echo -e "$INFO Finished $(date +%T--%d-%m-%Y)."
echo "----------------------------------------------------------------------------------------------------------------------------"

