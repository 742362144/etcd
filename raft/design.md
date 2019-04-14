## Progress

Progress represents a follower’s progress in the view of the leader. Leader maintains progresses of all followers, and sends `replication message` to the follower based on its progress. 

`replication message` is a `msgApp` with log entries.

A progress has two attribute: `match` and `next`. `match` is the index of the highest known matched entry. If leader knows nothing about follower’s replication status, `match` is set to zero. `next` is the index of the first entry that will be replicated to the follower. Leader puts entries from `next` to its latest one in next `replication message`.

A progress is in one of the three state: `probe`, `replicate`, `snapshot`. 

```
                            +--------------------------------------------------------+          
                            |                  send snapshot                         |          
                            |                                                        |          
                  +---------+----------+                                  +----------v---------+
              +--->       probe        |                                  |      snapshot      |
              |   |  max inflight = 1  <----------------------------------+  max inflight = 0  |
              |   +---------+----------+                                  +--------------------+
              |             |            1. snapshot success                                    
              |             |               (next=snapshot.index + 1)                           
              |             |            2. snapshot failure                                    
              |             |               (no change)                                         
              |             |            3. receives msgAppResp(rej=false&&index>lastsnap.index)
              |             |               (match=m.index,next=match+1)                        
receives msgAppResp(rej=true)                                                                   
(next=match+1)|             |                                                                   
              |             |                                                                   
              |             |                                                                   
              |             |   receives msgAppResp(rej=false&&index>match)                     
              |             |   (match=m.index,next=match+1)                                    
              |             |                                                                   
              |             |                                                                   
              |             |                                                                   
              |   +---------v----------+                                                        
              |   |     replicate      |                                                        
              +---+  max inflight = n  |                                                        
                  +--------------------+                                                        
```

When the progress of a follower is in `probe` state, leader sends at most one `replication message` per heartbeat interval. The leader sends `replication message` slowly and probing the actual progress of the follower. A `msgHeartbeatResp` or a `msgAppResp` with reject might trigger the sending of the next `replication message`.

当跟随者的进程处于“探测”状态时，leader每心跳间隔最多发送一条“复制消息”。领导者缓慢地发送“复制消息”，并探测追随者的实际进度。带有拒绝的' msgHeartbeatResp '或' msgAppResp '可能触发发送下一个'复制消息'。

When the progress of a follower is in `replicate` state, leader sends `replication message`, then optimistically increases `next` to the latest entry sent. This is an optimized state for fast replicating log entries to the follower.
当跟随者的进度处于“复制”状态时，leader发送“复制消息”，然后乐观地将“next”增加到发送的最新条目。这是将日志条目快速复制到关注者的优化状态。

When the progress of a follower is in `snapshot` state, leader stops sending any `replication message`.
当跟随者的进程处于“快照”状态时，leader停止发送任何“复制消息”。

A newly elected leader sets the progresses of all the followers to `probe` state with `match` = 0 and `next` = last index. The leader slowly (at most once per heartbeat) sends `replication message` to the follower and probes its progress.
新当选的领导人将所有追随者的进程设置为' probe '状态，' match ' = 0， ' next ' = last index。领导者缓慢地(每次心跳最多一次)向追随者发送“复制消息”，并探测其进程。

A progress changes to `replicate` when the follower replies with a non-rejection `msgAppResp`, which implies that it has matched the index sent. At this point, leader starts to stream log entries to the follower fast. The progress will fall back to `probe` when the follower replies a rejection `msgAppResp` or the link layer reports the follower is unreachable. We aggressively reset `next` to `match`+1 since if we receive any `msgAppResp` soon, both `match` and `next` will increase directly to the `index` in `msgAppResp`. (We might end up with sending some duplicate entries when aggressively reset `next` too low.  see open question)
当跟随者回复一个非拒绝的“msgAppResp”时，进程就会变为“复制”，这意味着它已经匹配了发送的索引。此时，leader开始快速地向follower发送日志条目。当跟随者回复一个拒绝“msgAppResp”或链路层报告跟随者不可用时，进程将回到“探针”。我们积极重置' next '为' match ' +1，因为如果我们很快收到任何' msgAppResp '， ' match '和' next '都会直接增加到' msgAppResp '中的' index '。(当“next”重置得太低时，我们可能会发送一些重复的条目。见悬而未决的问题)

A progress changes from `probe` to `snapshot` when the follower falls very far behind and requires a snapshot. After sending `msgSnap`, the leader waits until the success, failure or abortion of the previous snapshot sent. The progress will go back to `probe` after the sending result is applied.
当跟随者远远落后并且需要快照时，进度就从“探测”变为“快照”。发送“msgSnap”后，leader等待前面发送的快照的成功、失败或终止。应用发送结果后，进程将返回到“探测”。

### Flow Control

1. limit the max size of message sent per message. Max should be configurable.
Lower the cost at probing state as we limit the size per message; lower the penalty when aggressively decreased to a too low `next`
限制每个消息发送的最大消息大小。Max应该是可配置的。
降低探测状态的成本，因为我们限制了每个消息的大小;当攻击性降低到过低的“下一个”时，降低惩罚

2. limit the # of in flight messages < N when in `replicate` state. N should be configurable. Most implementation will have a sending buffer on top of its actual network transport layer (not blocking raft node). We want to make sure raft does not overflow that buffer, which can cause message dropping and triggering a bunch of unnecessary resending repeatedly.
当处于“复制”状态时，限制in flight消息的# < N。N应该是可配置的。大多数实现将在其实际网络传输层的顶部有一个发送缓冲区(而不是阻塞raft节点)。我们希望确保raft不会溢出该缓冲区，这可能导致消息丢失并触发大量不必要的重复重发。 
