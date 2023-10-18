#!/usr/bin/env bash

# Created Date: 2023-10-17
# Author: caoxianfei

g_binarypath="/usr/local/spdk/bin/spdk_tgt"
g_cpumask=0x03
g_sockname="/usr/local/spdk/spdk.sock"
g_rpcpath="/usr/local/spdk/script/rpc.py"
g_spdk_tgtd_log="/tmp/__curveadm_start_target__"

process_name="spdk_tgt"

umask 000
mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab

LD_LIBRARY_PATH=/lib ${g_binarypath} -m ${g_cpumask} -r ${g_sockname} > $g_spdk_tgtd_log 2>&1
if [ $? -ne 0 ]; then
	cat $g_spdk_tgtd_log
	exit 1
else 
	echo "SUCCESS"
fi
