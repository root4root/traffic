#!/bin/bash

NFLOG_GROUP="11"
TUN_DEVICE="vpn0"
IN_MARK="0x1"
OUT_MARK="0x2"

function start()
{
    /usr/sbin/iptables -I INPUT 1 ! -i $TUN_DEVICE -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -I OUTPUT 1 -m mark ! --mark $OUT_MARK -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -I FORWARD 1 -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -t mangle -I FORWARD 1 -i $TUN_DEVICE -j MARK --set-mark $IN_MARK
}

function stop()
{
    /usr/sbin/iptables -D INPUT  ! -i $TUN_DEVICE -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -D OUTPUT -m mark ! --mark $OUT_MARK -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -D FORWARD -j NFLOG --nflog-group $NFLOG_GROUP
    /usr/sbin/iptables -t mangle -D FORWARD -i $TUN_DEVICE -j MARK --set-mark $IN_MARK
}

case $1 in
"start")
    start
    ;;
"stop" )
    stop
    ;;
*)
    echo "Usage: start|stop"
    return 1
    ;;
esac
