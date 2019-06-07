#!/usr/bin/env bash

kod_db=/var/db/ntp-kod

if [[ ! -f "${kod_db}" ]]; then
    sudo touch "${kod_db}"
    sudo chmod 666 "${kod_db}"
fi

sudo sntp -sS time.apple.com
