#!/bin/bash

if [[ $1 == "path" && $2 != "" ]]; then
  version=$(curl -s -o- https://github.com/luizalabs/teresa/releases/latest | sed 's/.*tag\/\(.*\)\".*/\1/')

  curl -L -O "https://github.com/luizalabs/teresa/releases/download/$version/teresa-linux-amd64"

  chmod +x teresa-linux-amd64

  mv teresa-linux-amd64 "$2/teresa"

  printf "\n\e[1;33m[WARNING] \e[m\n"
  echo "Don't forget to add the specified \`path\` to your environment PATH."

  echo "You can do this by adding the code below to your ~/.bashrc:"
  
  echo "
  if [[ ! \"\$PATH\" == *$2*  ]]; then
    export PATH=\"\${PATH:+\${PATH}:}$2\"
  fi
  "
  echo "After that, restart your terminal :)"
elif [[ $1 == "" ]]; then
  if [[ $EUID -ne 0  ]]; then
    echo "If you don't specify \`path\` you need to run as root(sudo)" 
    exit 1
  fi

  version=$(curl -s -o- https://github.com/luizalabs/teresa/releases/latest | sed 's/.*tag\/\(.*\)\".*/\1/')

  curl -L -O "https://github.com/luizalabs/teresa/releases/download/$version/teresa-linux-amd64"

  chmod +x teresa-linux-amd64

  mv teresa-linux-amd64 /usr/local/bin/teresa
  printf "\n\e[0;32mInstallation success!\e[m\n"
fi
