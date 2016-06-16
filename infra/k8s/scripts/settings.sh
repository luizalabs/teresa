#!/bin/bash
# Name: recorevy-kube-master.sh
# Description: Recovery Kubernetes Master
# Autor: Eraldo S.Bernardino
# Date: 01/06/2016

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
MASTER_USER_DATA="/home/oregon/us-west-2-kubernentes/master-user-data.gz"
MASTER_INSTANCE_TYPE="m3.medium"
MASTER_KEY_NAME="kubernetes-3a0ab4c58a85f9173c85e7aaf2df8baa"
MASTER_AMI_ID="ami-7840ac18"
LOG="recovery-kube-master.log"
ENVIRONMENT="Test"
#ENVIRONMENT="staging"
#ENVIRONMENT="production"
