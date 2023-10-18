#!/usr/bin/env bash

# Created Date: 2023-10-17
# Author: caoxianfei

g_hostname=$1
g_cachesize=$2
g_hugepage=$3

g_binarypath="/usr/local/spdk/bin/spdk_tgt"
g_cpumask=0x03
g_sockname="/usr/local/spdk/spdk.sock"
g_rpcpath="/usr/local/spdk/scripts/rpc.py"
g_setuppath="/usr/local/spdk/scripts/setup.sh"
g_spdk_tgtd_log="/tmp/__curveadm_set_target__"

process_name="spdk_tgt"

HUGE_EVEN_ALLOC=yes HUGEMEM=$g_hugepage PCI_ALLOWED="none" ${g_setuppath} > $g_spdk_tgtd_log 2>&1
if [ $? -ne 0 ]; then
 echo "HUGE_EVEN_ALLOC failed"
 cat $g_spdk_tgtd_log
 exit 1
fi

# create memdisk
memdisk=Malloc_host_
${g_rpcpath} -s ${g_sockname} bdev_malloc_create -b $memdisk $g_cachesize 512 > $g_spdk_tgtd_log 2>&1
if [ $? -ne 0 ]; then
 echo "bdev_malloc_create failed"
 cat $g_spdk_tgtd_log
 exit 1
fi

# add portal group 
${g_rpcpath} -s ${g_sockname} iscsi_create_portal_group 1 ${g_hostname}:3260 > $g_spdk_tgtd_log 2>&1
if [ $? -ne 0 ]; then
 echo "add a portal group failed"
 cat $g_spdk_tgtd_log
 exit 1
fi

# create initiator group
${g_rpcpath} -s ${g_sockname} iscsi_create_initiator_group 2 ANY 10.0.0.0/8 > $g_spdk_tgtd_log 2>&1
if [ $? -ne 0 ]; then
 echo "add an initiator group failed"
 cat $g_spdk_tgtd_log
 exit 1
fi

