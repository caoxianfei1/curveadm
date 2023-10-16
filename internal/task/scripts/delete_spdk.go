/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2023-10-16
 * Author: caoxianfei1
 */

package scripts

var DELETE_SPDK_TARGET = `
#!/usr/bin/env bash

g_volume=$1

g_sockname="/tmp/spdk.sock"
g_rpcpath="/curvebs/spdk/scripts/rpc.py"
g_spdk_log="./spdk.log"

target=iqn.2016-06.io.spdk:${g_volume}
sudo ${g_rpcpath} -s ${g_sockname} iscsi_delete_target_node $target >> $g_spdk_log 2>&1
if [ $? -ne 0 ]; then
	echo "delete target node failed"
	exit 1
fi

ocf=ocf_${g_volume}
sudo ${g_rpcpath} -s ${g_sockname} bdev_ocf_delete $ocf >> $g_spdk_log 2>&1
if [ $? -ne 0 ]; then
	echo "delete ocf failed"
	exit 1
fi

cbd_bdev=cbd_${g_volume}
sudo ${g_rpcpath} -s ${g_sockname} bdev_cbd_delete $cbd_bdev >> $g_spdk_log 2>&1
if [ $? -ne 0 ]; then
	echo "delete cbd dev failed"
	exit 1
fi

memdisk=Malloc_${g_volume}
sudo ${g_rpcpath} -s ${g_sockname} bdev_malloc_delete $memdisk >> $g_spdk_log 2>&1
if [ $? -ne 0 ]; then
	echo "delete malloc memory disk failed"
	exit 1
fi
`
