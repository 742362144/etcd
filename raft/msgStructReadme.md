'MsgHup'用于选举。如果节点是追随者或候选者，那么
'raft'结构中的'tick'函数设置为'tickElection'。如果是追随者或
候选人在选举超时之前没有收到任何心跳，它
将'MsgHup'传递给其Step方法并成为（或保持）候选者
开始新的选举。

'MsgBeat'是一种内部类型，表示领导者发送心跳
'MsgHeartbeat'类型。如果一个节点是一个领导者，那么'tick'函数就是
'raft'结构设置为'tickHeartbeat'，并触发领导者
定期向其关注者发送“MsgHeartbeat”消息。

'MsgProp'建议将数据附加到其日志条目中。这是一个特殊的
键入以将提案重定向到领导者。因此，send方法会被覆盖
raftpb.Message的术语，其HardState的术语，以避免附加它
当地术语'MsgProp'。当'MsgProp'被传递给领导者的'Step'时
方法，领导者首先调用'appendEntry'方法来追加条目
到它的日志，然后调用'bcastAppend'方法发送这些条目
它的同行。传递给候选人时，'MsgProp'被删除。传递给
关注者，'MsgProp'由发送存储在关注者的邮箱（msgs）中
方法。它与发件人的ID一起存储，然后转发给leader
rafthttp包。

'MsgApp'包含要复制的日志条目。一位领导人打电话给bcastAppend，
调用sendAppend，它会在'MsgApp'中发送即将复制的日志
类型。当'MsgApp'传递给候选人的步骤方法时，候选人会恢复
回到追随者，因为它表明有一个有效的领导者发送
'MsgApp'消息。候选人和追随者回复此消息
'MsgAppResp'类型。

'MsgAppResp'是对日志复制请求（'MsgApp'）的响应。什么时候
'MsgApp'被传递给候选人或追随者的Step方法，它会响应
调用'handleAppendEntries'方法，将'MsgAppResp'发送到raft
邮箱。

'MsgVote'要求投票选举。当节点是追随者或
候选者和'MsgHup'被传递给它的Step方法，然后节点调用
“竞选”方法让竞选活动成为领导者。一旦'广告系列'
调用方法，节点成为候选节点并将“MsgVote”发送给对等节点
在群集中请求投票。当传递给领导者或候选人的步骤时
方法和消息的期限低于领导者或候选人，
'MsgVote'将被拒绝（'MsgVoteResp'返回时Reject为true）。
如果领导者或候选人获得更高期限的“MsgVote”，它将会恢复
回到追随者。当'MsgVote'传递给追随者时，它投票支持
发件人仅在发件人的最后一个期限大于MsgVote的期限或时
发件人的最后一个术语等于MsgVote的术语，但发件人的最后一个承诺
index大于或等于跟随者。

'MsgVoteResp'包含来自投票请求的回复。当'MsgVoteResp'是
传递给候选人，候选人计算它赢得了多少票。如果
它超过了大多数（法定人数），它成为领导者并称之为'bcastAppend'。
如果候选人获得多数否决票，则会恢复原状
追随者。

'MsgPreVote'和'MsgPreVoteResp'用于可选的两阶段选举
协议。当Config.PreVote为true时，首先进行预选
（使用与常规选举相同的规则），没有节点增加其术语
数字，除非选举前表明竞选节点将获胜。
当分区节点重新加入群集时，这可以最大限度地减少中断。

'MsgSnap'请求安装快照消息。当一个节点刚刚
它呼吁，成为领导者或领导者收到'MsgProp'消息
'bcastAppend'方法，然后将'sendAppend'方法调用到每个方法
追随者。在'sendAppend'中，如果领导者未能获得条款或条目，
领导者通过发送'MsgSnap'类型消息来请求快照。

'MsgSnapStatus'告诉快照安装消息的结果。当一个
追随者拒绝'MsgSnap'，它表示快照请求
'MsgSnap'因导致网络层的网络问题而失败
无法向其关注者发送快照。然后领导考虑
追随者作为探索的进展。当'MsgSnap'没有被拒绝时，它
表示快照成功，领导者设置了跟随者
探测并恢复其日志复制的进度。

'MsgHeartbeat'从领导者那里发出心跳。当'MsgHeartbeat'通过时
候选人和消息的期限高于候选人的候选人
恢复追随者并更新其提交的索引
这个心跳。它将消息发送到其邮箱。什么时候
'MsgHeartbeat'被传递给跟随者的Step方法，而消息的术语是
跟随者高于追随者，用ID更新其leaderID
从消息。

'MsgHeartbeatResp'是对'MsgHeartbeat'的回复。什么时候'MsgHeartbeatResp'
传递给领导者的Step方法，领导者知道哪个粉丝
回应。只有当领导者最后提交的指数大于
跟随者的匹配索引，领导者运行'sendAppend`方法。

'MsgUnreachable'告诉我们没有提供请求（消息）。什么时候
领导者发现'MsgUnreachable'被传递给领导者的Step方法
通常，无法访问发送此“MsgUnreachable”的关注者
表示'MsgApp'丢失了。当追随者的进步状态复制时，
领导者将其重新开始探测。