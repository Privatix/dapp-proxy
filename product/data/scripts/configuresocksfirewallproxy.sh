#!/bin/sh

store_file="$2"

if [ "$1" == "on" ]; then
    if [ -f "$store_file" ]; then
        echo "file '$store_file' exists" && exit 1
    fi

    host=$3
    port=$4

    networksetup -listallnetworkservices|grep -v "(*)"| while read in
    do 
        service="$in"
        echo "$service" >> "$store_file"
        echo networksetup -setsocksfirewallproxy "$service" "$host" "$port"
        networksetup -setsocksfirewallproxy "$service" "$host" "$port"
        echo networksetup -setsocksfirewallproxystate "$service" on
        networksetup -setsocksfirewallproxystate "$service" on
    done
else
    if [ ! -f "$store_file" ]; then
        echo "file '$store_file' does not exist" && exit 1
    fi

    cat "$store_file" | while read in
    do
        service="$in"
        echo networksetup -setsocksfirewallproxystate "$service" off
        networksetup -setsocksfirewallproxystate "$service" off
    done

    rm "$store_file"
fi
