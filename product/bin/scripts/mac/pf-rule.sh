#!/bin/sh

if [ -n "$1" ]
then
	status=$1
else
	echo "No arguments supplied (status)."
	exit 1
fi

if [ "$status" = "on" ]
then

    if [ -n "$2" ]
    then
	    port=$2
    else
	    echo "No arguments supplied (port)."
	    exit 1
    fi

    if [ -n "$3" ]
    then
	    rulefile=$3
    else
	    echo "No arguments supplied (rulefile)."
	    exit 1
    fi

    # creates rules
    rm -f "$rulefile"

    echo "pass in proto { tcp, udp } from any to any port $port" >> "$rulefile"

    #disables pfctl
    /sbin/pfctl -d
    sleep 1

    #flushes all pfctl rules
    /sbin/pfctl -F all
    sleep 1

    #starts pfctl and loads the rules from the rules file
    /sbin/pfctl -f "$rulefile" -e
elif [ "$status" = "off" ]
then
    #disables pfctl
    /sbin/pfctl -d
    sleep 1

    #flushes all pfctl rules
    /sbin/pfctl -F all
    sleep 1

    #starts pfctl and loads the default rules
    /sbin/pfctl -f /etc/pf.conf -e
fi
