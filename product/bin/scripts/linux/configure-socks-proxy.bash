#!/bin/bash

store_file="$2"

if [ "$1" == "on" ]; then

    # backup current configuration.
    echo "$(gsettings get org.gnome.system.proxy mode);$(gsettings get org.gnome.system.proxy.socks port);$(gsettings get org.gnome.system.proxy.socks host)" > "$store_file"

    port=$3
    if [ -f "$port" ]; then
        echo "port not specified" && exit 1
    fi

    gsettings set org.gnome.system.proxy mode 'manual'
    gsettings set org.gnome.system.proxy.socks port $port
    gsettings set org.gnome.system.proxy.socks host 'localhost'
else
    # restore configuration from backup.
    IFS=";" read -r -a content <<< $(cat "$store_file");
    gsettings set org.gnome.system.proxy mode "${content[0]}"
    gsettings set org.gnome.system.proxy.socks port "${content[1]}"
    gsettings set org.gnome.system.proxy.socks host "${content[2]}"
fi
