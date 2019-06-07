#!/bin/bash

if [ -n "$1" ]
then
	status=$1
else
	echo "No arguments supplied (status)."
	exit 1
fi

if [ -n "$2" ]
then
    port=$2
else
    echo "No arguments supplied (port)."
fi



if [ "$status" = "on" ]
then
    iptables -A INPUT -p tcp --dport "$port" -j ACCEPT
    iptables -A INPUT -p udp --dport "$port" -j ACCEPT
elif [ "$status" = "off" ]
then
    iptables -D INPUT -p tcp --dport "$port" -j ACCEPT
    iptables -D INPUT -p udp --dport "$port" -j ACCEPT
fi
