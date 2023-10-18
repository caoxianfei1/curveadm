/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-04-04
* Author: wanghai (SeanHai)
 */

package manager

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command/target"
	"github.com/opencurve/curveadm/http/core"
	"github.com/opencurve/pigeon"
)

const (
	KEY_ETCD_ENDPOINT                 = "etcd.address"
	KEY_MDS_ENDPOINT                  = "mds.address"
	KEY_MDS_DUMMY_ENDPOINT            = "mds.dummy.address"
	KEY_SNAPSHOT_CLONE_DUMMY_ENDPOINT = "snapshot.clone.dummy.address"
	KEY_SNAPSHOT_CLONE_PROXY_ENDPOINT = "snapshot.clone.proxy.address"
	KEY_MONITOR_PROMETHEUS_ENDPOINT   = "monitor.prometheus.address"
)

func newAdmFail(r *pigeon.Request, err error) bool {
	r.Logger().Error("failed when new curveadm",
		pigeon.Field("error", err))
	return core.Exit(r, err)
}

// spdk target
func StartSpdkTgtdHandler(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*StartSpdkTgtdRequest)
	err = target.StartTgtd(adm, data.Host, data.Cache, data.HugePageMem, data.Client)
	if err != nil {
		r.Logger().Error("StartSpdkTgtdHandler failed",
			pigeon.Field("host name", data.Host),
			pigeon.Field("client config", data.Client),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func AddSpdkTgtHandler(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}

	data := ctx.Data.(*AddSpdkTgtRequest)
	err = target.AddSpdkTgt(adm,
		data.Image,
		data.Host,
		data.Size,
		data.CacheSize,
		data.BlockSize,
		data.Create,
		data.CreateCache,
	)

	if err != nil {
		r.Logger().Error("AddSpdkTgtHandler failed",
			pigeon.Field("host name", data.Host),
			pigeon.Field("image", data.Image),
			pigeon.Field("image size", data.Size),
			pigeon.Field("block size", data.BlockSize),
			pigeon.Field("cache size", data.CacheSize),
			pigeon.Field("create image?", data.Create),
			pigeon.Field("create cache disk", data.CreateCache),

			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func DeleteSpdkTgtHandler(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*DeleteSpdkTgtRequest)
	err = target.DeleteSpdkTgt(adm, data.Target, data.Host)
	if err != nil {
		r.Logger().Error("DeleteSpdkTgtHandler failed",
			pigeon.Field("host name", data.Host),
			pigeon.Field("target", data.Target),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func ListSpdkTgtHandler(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*ListSpdkTgtRequest)
	tgts, err := target.ListSpdkTgt(adm, data.Host)
	if err != nil {
		r.Logger().Error("ListSpdkTgtHandler failed",
			pigeon.Field("host name", data.Host),
			pigeon.Field("error", err))
	}
	return core.ExitSuccessWithData(r, tgts)
}

func StopSpdkTgtdHandler(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*StopSpdkTgtdRequest)
	err = target.StopSpdkTgtd(adm, data.Host)
	if err != nil {
		r.Logger().Error("StopSpdkTgtdHandler failed",
			pigeon.Field("host name", data.Host),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}
