#!/usr/bin/env bash

# Usage: target USER VOLUME CREATE SIZE
# Example: target curve test true 10
# See Also: https://linux.die.net/man/8/tgtadm
# Created Date: 2022-02-08
# Author: Jingli Chen (Wine93)


g_user=$1
g_volume=$2
g_create=$3
g_size=$4
g_blocksize=$5
g_cachesize=$6
g_spdk=$7
g_hostname=$8
g_create_cache=$9

g_tid=1
g_sockname="/usr/local/spdk/spdk.sock"
g_rpcpath="/usr/local/spdk/scripts/rpc.py"
g_spdk_log="/tmp/__curveadm_add_target__"

g_image=cbd:pool/${g_volume}_${g_user}_
g_image_md5=$(echo -n ${g_image} | md5sum | awk '{ print $1 }')
g_targetname=iqn.$(date +"%Y-%m").com.opencurve:curve.${g_image_md5}

output_file=/tmp/spdk_output.txt

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab

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
		vol=$(echo $line | awk -F ':' '{print $3}')
  		if [[ "$vol" == "$2" ]]; then
			return 0
  		fi
	done <<< "$output"

	return 1
}

if [ $g_create == "true" ]; then
    output=$(curve_ops_tool create -userName=$g_user -fileName="${g_volume}" -fileLength=$g_size)
    if [ $? -ne 0 ]; then
        if [ "$output" != "CreateFile fail with errCode: kFileExists" ]; then
			echo $output
            exit 1
        fi
    fi
fi

if [ $g_spdk == "false" ]; then
	for ((i=1;;i++)); do
	    tgtadm --lld iscsi --mode target --op show --tid $i 1>/dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        g_tid=$i
	        break
	    fi
	done

	tgtadm --lld iscsi \
	   --mode target \
	   --op new \
	   --tid ${g_tid} \
	   --targetname ${g_targetname}
	if [ $? -ne 0 ]; then
	   echo "tgtadm target new failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode logicalunit \
	    --op new \
	    --tid ${g_tid} \
	    --lun 1 \
	    --bstype curve \
	    --backing-store ${g_image} \
	    --blocksize ${g_blocksize}
	if [ $? -ne 0 ]; then
	   echo "tgtadm logicalunit new failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode logicalunit \
	    --op update \
	    --tid ${g_tid} \
	    --lun 1 \
	    --params vendor_id=NetEase,product_id=CurveVolume,product_rev=2.0
	if [ $? -ne 0 ]; then
	   echo "tgtadm logicalunit update failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode target \
	    --op bind \
	    --tid ${g_tid} \
	    -I ALL
	if [ $? -ne 0 ]; then
	   echo "tgtadm target bind failed"
	   exit 1
	fi
else
	volume=${g_volume//\//_} # volume=_test1_test2
	target_volume=${g_volume//\//-} # target_volume=-test1-test2
	target_name=${target_volume#[^a-zA-Z]} # target_name=test1-test2
	# one target refer to one volume
	checkTgtExist $g_hostname $target_name
	if [ $? -eq 0 ]; then
		echo "EXIST"
		exit 0
	fi
	
	declare memdisk
	if [ "$g_create_cache" == "true" ]; then
		memdisk=Malloc${volume}
		${g_rpcpath} -s ${g_sockname} bdev_malloc_create -b $memdisk $g_cachesize 512 > $g_spdk_log 2>&1
		if [ $? -ne 0 ]; then
			echo " bdev_malloc_create execution failed"
			cat $g_spdk_log
			exit 1
		fi
	else 
		memdisk=Malloc_host_
	fi

	cbd_path=/${g_volume}_${g_user}_
	cbd_bdev=cbd${volume}
	${g_rpcpath} -s ${g_sockname} bdev_cbd_create -b $cbd_bdev --cbd $cbd_path --exclusive=0 --blocksize=4096 > $g_spdk_log 2>&1
	if [ $? -ne 0 ]; then
		echo "bdev_cbd_create execution failed"
		cat $g_spdk_log
		exit 1
	fi
	ocf=ocf${volume}
	${g_rpcpath} -s ${g_sockname} bdev_ocf_create $ocf wb $memdisk $cbd_bdev > $g_spdk_log 2>&1
	if [ $? -ne 0 ]; then
		echo "bdev_ocf_create execution failed"
		cat $g_spdk_log
		exit 1
	fi

	lun_pair=${ocf}:0
	${g_rpcpath} -s ${g_sockname} iscsi_create_target_node $target_name $target_name $lun_pair 1:2 1024 -d > $g_spdk_log 2>&1
	if [ $? -ne 0 ]; then
		echo "iscsi_create_target_node execution failed"
		cat $g_spdk_log
		exit 1
	fi
	
	checkTgtExist $g_hostname $target_name
	if [ $? -ne 0 ]; then
		echo "add target $target_name failed"
		exit 1
	fi
fi


