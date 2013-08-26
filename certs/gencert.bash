#!/usr/bin/env bash

cdir=$( cd "$( dirname "$0" )" && pwd )
conf="$cdir/openssl.cnf"

out='.'
ca="$out/ca"
serial="$out/serial"

case $1 in
	ca)
		echo 'certificate authority'
		openssl req -x509 -newkey rsa:2048 -keyout $ca.key -out $ca.crt -config $conf -extensions v3_ca \
		&& chmod 400 $ca.key
		;;
	server)
		echo 'server certificate'
		echo "if the server has more than one valid domain or IP, specify these in 'subjectAltName' in $conf"
		echo "by default 'subjectAltName' is set to 'IP:127.0.0.1,DNS:localhost' for local testing"
		read -p 'Enter primary server domain or IP: ' domain
		srv="$out/server"
		subj=$( openssl x509 -in $ca.crt -noout -subject | sed 's/subject=\s*//' )
		openssl req -new -newkey rsa:2048 -nodes -keyout $srv.key -out $srv.csr -config $conf -subj "$subj/CN=$domain" \
		&& chmod 400 $srv.key \
		&& openssl x509 -req -in $srv.csr -out $srv.crt -extfile $conf -extensions v3_srv \
		   -CA $ca.crt -CAkey $ca.key -CAserial $serial -CAcreateserial \
		&& rm $srv.csr
		;;
	client)
		echo 'client certificate'
		read -p 'Enter user name: ' username
		cli="$out/$username"
		subj=$( openssl x509 -in $ca.crt -noout -subject | sed 's/subject=\s*//' )
		openssl req -new -newkey rsa:2048 -nodes -keyout $cli.key -out $cli.csr -config $conf -subj "$subj/CN=$username" \
		&& chmod 400 $cli.key \
		&& openssl x509 -req -in $cli.csr -out $cli.crt -extfile $conf -extensions v3_cli \
		   -CA $ca.crt -CAkey $ca.key -CAserial $serial -CAcreateserial \
		&& rm $cli.csr \
		&& openssl pkcs12 -export -in $cli.crt -inkey $cli.key -certfile $ca.crt -out $cli.p12 \
		&& chmod 400 $cli.p12
		;;
	*)
		echo "usage: $0 command"
		echo "command can be:"
		echo "	ca     - generate a certificate authority key and crt file"
		echo "	server - generate a server certificate key and crt file"
		echo "	client - generate a client certificate key, crt and p12 file"
		exit 1
		;;
esac

