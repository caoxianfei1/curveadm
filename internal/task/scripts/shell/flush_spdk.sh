#!/bin/bash

# Created Date: 2023-11-20
# Author: caoxianfei

# g_target=iqn.2016-06.io.spdk:test2
g_target=$1
g_hostname=$2

g_sockname="/usr/local/spdk/spdk.sock"
g_rpcpath="/usr/local/spdk/scripts/rpc.py"
g_spdk_log="/tmp/__curveadm_flush_target__"

g_volume=$(echo $g_target | awk -F ':' '{print $2}') # g_volume=test1-test2
g_volume=/${g_volume//-/\/} # g_volume=/test1/test2

volume=${g_volume//\//_} # volume=_test1_test2

output_file=/tmp/spdk_output.txt
function checkTgtExist(){
	iscsiadm --mode discovery -t sendtargets --portal $1:3260 > $output_file 2>&1
	if [ $? -ne 0 ]; then
		output=$(cat $output_file)
		if [ "$output" == "iscsiadm: No portals found" ]; then
			return 1
		else
			echo $output
			exit 1
		fi
	fi
	output=$(cat $output_file)
	while read -r line; do
		target=$(echo $line | awk '{print $2}')
  		if [[ "$target" == "$2" ]]; then
			return 0
  		fi
	done <<< "$output"

	return 1
}

checkTgtExist $g_hostname $g_target
if [ $? -eq 1 ]; then
	echo "target does not exist"
	exit 1
fi

ocf=ocf${volume}
${g_rpcpath} -s ${g_sockname} bdev_ocf_flush $ocf > $g_spdk_log 2>&1
if [ $? -ne 0 ]; then
	echo "bdev_ocf_flush execution failed"
	cat $g_spdk_log
	exit 1
fi