#!/bin/bash
# Name: recorevy-kube-master.sh
# Description: Recovery Kubernetes Master
# Autor: Eraldo S.Bernardino
# Date: 01/06/2016
AWS_DEFAULT_OUTPUT="${AWS_DEFAULT_OUTPUT:-text}"

#Getting Global Variables
source settings.sh
source kubernetes-helpers.sh
#Log
LOG="recovery-kube-master.log"
banner() {
    msg="# $* #"
    edge=$(echo -e "$msg" | sed 's/./#/g')
    echo -e "$edge"
    echo -e "$msg"
    echo -e "$edge"
 }
banner "                                      Kubernetes Recovery Master  ver.1.0                                               "
echo -e $CYAN"                                                Starting process..."$NONE
VPC_ID=$(get_vpc_id)
MASTER_EBS_ID=$(get_master_ebs_id)
MASTER_ID=$(get_master_id $VPC_ID)
MASTER_EIP=$(get_master_eip $MASTER_EBS_ID)

echo  "----------------------------------------------------------------------------------------------------------------------------"
echo -e "$INFO Starting $(date +%T--%d-%m-%Y)."
echo -e "$INFO AWS Zone is $GREEN$ZONE$NONE."
echo -e "$INFO AWS Availability Zone is $GREEN$AZ$NONE."
echo -e "$INFO AWS VPC-ID is $GREEN$VPC_ID$NONE."
echo -e "$INFO Kubernetes ID is $GREEN$CLUSTER_ID$NONE."
echo -e "$INFO Kubernetes Master Name is $GREEN$MASTER_NAME$NONE."

if [[ -n "${MASTER_ID}" ]] ; then
    MASTER_STATE=$(get_master_state $MASTER_ID)
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
    EIP_STATE=$(get_eip_state $MASTER_EIP)
    if [[ -n "${EIP_STATE}" ]]; then
            echo -e "$ERRO Elastic IP $RED$MASTER_EIP$NONE allocated for InstanceID = $RED$EIP_STATE$NONE."

    else
        echo -e "$INFO ElasticIP $GREEN$MASTER_EIP$NONE is available."
        PIP_STATE=$(get_master_pip)

        if [[ -n "${PIP_STATE}" ]]; then
            echo -e "$ERRO Private IP $RED$MASTER_PIP$NONE allocated for InstanceID = $RED$PIP_STATE$NONE."
        else
            echo -e "$INFO Private IP $GREEN$MASTER_PIP$NONE is available."
            echo -e "$INFO Starting create new kubernetes master."
            MASTER_ID=$(create_kube_master $VPC_ID)
            wait_for_instance_state $MASTER_ID "running"
            get_eip_allocationid $MASTER_EIP
            exec_associate_eip $MASTER_ID $eip_allocationid $MASTER_EIP
            exec_attach_volume $MASTER_ID $MASTER_EBS_ID
            create_master_tags $MASTER_ID $ENVIRONMENT
        fi
    fi
fi
echo -e "$INFO Finished $(date +%T--%d-%m-%Y)."
echo "----------------------------------------------------------------------------------------------------------------------------"

