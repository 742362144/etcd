// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"strings"

	"go.etcd.io/etcd/raft/raftpb"
)

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	// client的数据变更请求
	proposeC := make(chan string)
	defer close(proposeC)
	// client的配置更改请求
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var kvs *kvstore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }

	// newRaftNode initiates a raft instance
	// commitC  :  committed log entry channel
	// errorC   :  error channel
	// proposeC :  Proposals for log updates are sent over the provided the proposal channel.
	// All log entries are replayed over the commit channel, followed by a nil message (to indicate the channel is
	// current), then new log entries.
	// To shutdown, close proposeC and read errorC.
	commitC, errorC, snapshotterReady := newRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	// 新建KVStore
	// KVStore可以看做一个map
	// 任务很简单，就是接受commitC传来的已提交日志，将其应用到这个map上
	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	// the key-value http handler will propose updates to raft
	// serveHttpKVAPI starts a key-value server with a GET/PUT API and listens.
	// confChangeC : client用来提交集群配置请求给raft模块
	// errorC      : 当raft出错或要退出时，关闭HttpServer
	serveHttpKVAPI(kvs, *kvport, confChangeC, errorC)
}
