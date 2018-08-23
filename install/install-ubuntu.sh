#!/usr/bin/env bash

# Installing dependencies
echo "[ DEPENDENCIES CHECK ]"
if [[ $(dpkg -s curl | grep -Po '(?<=Status: )[^$]*') != "install ok installed" ]]; then
    echo " - installing dependency: 'curl'"
    sudo apt-get install curl
else
    echo " - dependency 'curl' already installed"
fi

# Getting latest release download URI

echo "[ LATEST DOWNLOAD URI CHECK ]"
LATEST_URI=https://github.com/luizalabs/teresa/releases/latest/
LATEST_VERSION=$(curl $LATEST_URI -s -L -I -o /dev/null -w '%{url_effective}' | grep -Po '(?<=tag\/)[^$]*')
DOWNLOAD_URI=https://github.com/luizalabs/teresa/releases/download/$LATEST_VERSION/teresa-linux-amd64
echo " - download URI: '$DOWNLOAD_URI'"

# Downloading and Installing
echo "[ DOWNLOADING & INSTALLING ]"
echo " - downloading"
sudo wget $DOWNLOAD_URI -xO /opt/teresa/teresa -q
sudo chmod +x /opt/teresa/teresa
echo " - checking 'teresa'"

if [[ $(stat -c "%a" /opt/teresa/teresa) == 755 ]]; then
    echo "   - status: OK"
else
    echo "   - status: ERROR"
fi

if [[ -z $(cat ~/.bashrc | grep "export PATH=$PATH:/opt/teresa/") ]]; then
    echo " - inserting command to '.bashrc'"
    # Only if the command isn't written into .bashrc
    echo "# Enable 'teresa' command" >> ~/.bashrc
    echo "export PATH=$PATH:/opt/teresa/" >> ~/.bashrc
else:
    echo " - command already inserted to '.bashrc'"
fi;

echo " - enabling 'teresa' command"
source ~/.bashrc
echo "[ DONE ]"
