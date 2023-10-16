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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: target USER VOLUME CREATE SIZE
 * Example: target curve test true 10
 * See Also: https://linux.die.net/man/8/tgtadm
 */

var START_SPDK = `
#!/usr/bin/env bash

g_binarypath="/usr/local/spdk/bin/spdk_tgt"
g_cpumask=0x03
g_sockname="/tmp/spdk.sock"
g_iscsi_log="./spdk_tgtd.log"
process_name="spdk_tgt"

g_sockname="/tmp/spdk.sock"
g_rpcpath="/curvebs/spdk/scripts/rpc.py"

umask 000
mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab

if ps aux | grep -v grep | grep "$process_name" > /dev/null; then
   echo "spdk iscsi_tgt has already been started, now exit!"
   exit 0
fi

sudo LD_LIBRARY_PATH=/lib ${g_binarypath} -m ${g_cpumask} -r ${g_sockname} > $g_iscsi_log 2>&1
if ps aux | grep -v grep | grep "$process_name" > /dev/null; then
    echo "spdk tgt started success!"
else
    echo "spdk tgt started failed!"
    exit 1
fi

sudo ${g_rpcpath} -s ${g_sockname} iscsi_create_portal_group 1 ${g_host}:3260 > $g_iscsi_log 2>&1
if [ $? -ne 0 ]; then
	echo "add a portal group failed"
	exit 1
fi

sudo ${g_rpcpath} -s ${g_sockname} iscsi_create_initiator_group 2 ANY 10.0.0.0/8 >> $g_iscsi_log 2>&1
if [ $? -ne 0 ]; then
	echo "add an initiator group failed"
	exit 1
fi
`
