## kubernetes kubelet 触发清理worknode 上的无用镜像和容器(垃圾回收)

# Prequirements

因为需要对节点身上的`kubelet` 进程进行重新配置，需要重启`kubelet` 守护进程,所以提前得把上面的pod
驱逐到其他节点上，并且在集群中标记为不再被调度，等一切完毕后再加入集群，具体做法可参考：

https://blog.csdn.net/stonexmx/article/details/73543185





Reference


- https://blog.csdn.net/shida_csdn/article/details/99734411
- 
