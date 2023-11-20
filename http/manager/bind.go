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

import "github.com/opencurve/pigeon"

var METHOD_REQUEST map[string]Request

type (
	HandlerFunc func(r *pigeon.Request, ctx *Context) bool

	Context struct {
		Data interface{}
	}

	Request struct {
		httpMethod string
		method     string
		vType      interface{}
		handler    HandlerFunc
	}
)

func init() {
	METHOD_REQUEST = map[string]Request{}
	for _, request := range requests {
		METHOD_REQUEST[request.method] = request
	}
}

// spdk target
type StartSpdkTgtdRequest struct {
	Host        string `json:"host" binding:"required"`
	Cache       string `json:"cache"`
	HugePageMem string `json:"hugePageMem"`
	Client      string `json:"client"`
}

type AddSpdkTgtRequest struct {
	Image       string `json:"image"`
	Host        string `json:"host"`
	Create      bool   `json:"create"` // create volume or not
	Size        string `json:"size"`   // volume size
	CreateCache bool   `json:"createCache"`
	CacheSize   string `json:"cacheSize"`
	BlockSize   string `json:"blockSize"`
	WritePolicy string `json:"writePolicy"`
}

type DeleteSpdkTgtRequest struct {
	Target string `json:"target"`
	Host   string `json:"host"`
}

type ListSpdkTgtRequest struct {
	Host string `json:"host"`
}

type StopSpdkTgtdRequest struct {
	Host string `json:"host"`
}

var requests = []Request{
	// {
	// 	"GET",
	// 	"host.list",
	// 	ListHostRequest{},
	// 	ListHost,
	// },
	// {
	// 	"POST",
	// 	"host.commit",
	// 	CommitHostRequest{},
	// 	CommitHost,
	// },
	// {
	// 	"GET",
	// 	"disk.list",
	// 	ListDiskRequest{},
	// 	ListDisk,
	// },
	// {
	// 	"POST",
	// 	"disk.commit",
	// 	CommitDiskRequest{},
	// 	CommitDisk,
	// },
	// {
	// 	"GET",
	// 	"disk.format.status",
	// 	GetFormatStatusRequest{},
	// 	GetFormatStatus,
	// },
	// {
	// 	"GET",
	// 	"disk.format",
	// 	FormatDiskRequest{},
	// 	FormatDisk,
	// },
	// {
	// 	"GET",
	// 	"config.show",
	// 	ShowConfigRequest{},
	// 	ShowConfig,
	// },
	// {
	// 	"POST",
	// 	"config.commit",
	// 	CommitConfigRequest{},
	// 	CommitConfig,
	// },
	// {
	// 	"GET",
	// 	"cluster.list",
	// 	ListClusterRequest{},
	// 	ListCluster,
	// },
	// {
	// 	"POST",
	// 	"cluster.add",
	// 	AddClusterRequest{},
	// 	AddCluster,
	// },
	// {
	// 	"POST",
	// 	"cluster.checkout",
	// 	CheckoutClusterRequest{},
	// 	CheckoutCluster,
	// },
	// {
	// 	"GET",
	// 	"cluster.deploy",
	// 	DeployClusterRequest{},
	// 	DeployCluster,
	// },
	// {
	// 	"GET",
	// 	"cluster.service.addr",
	// 	GetClusterServicesAddrRequest{},
	// 	GetClusterServicesAddr,
	// },
	{
		"POST",
		"target.start",
		StartSpdkTgtdRequest{},
		StartSpdkTgtdHandler,
	},
	{
		"POST",
		"target.add",
		AddSpdkTgtRequest{},
		AddSpdkTgtHandler,
	},
	{
		"POST",
		"target.delete",
		DeleteSpdkTgtRequest{},
		DeleteSpdkTgtHandler,
	},
	{
		"POST",
		"target.list",
		ListSpdkTgtRequest{},
		ListSpdkTgtHandler,
	},
	{
		"POST",
		"target.stop",
		StopSpdkTgtdRequest{},
		StopSpdkTgtdHandler,
	},
}
