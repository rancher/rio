#!/bin/bash
while true; do
    issuer=$(openssl x509 -in $1 -text -noout | grep Issuer)
    if [[ $issuer =~ .*Encrypt.* ]]
    then
        exit 0
    fi
    sleep 1
done
