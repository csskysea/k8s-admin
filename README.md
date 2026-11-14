# k8s-yaml

一些k8s 应用部署yaml文件



一、 Pod 常见状态与排障

Pod 是 K8s 中最常见的故障单元。首先执行基础查询：
bash
kubectl get pods -A -o wide
kubectl describe pod <pod-name> -n <namespace>
请谨慎使用此类代码。
• 核心动作：重点看 Describe 输出底部的 Events（事件）记录，这是定位 90% 应用与调度问题的关键。

Pod 状态	常见原因	排障命令 / 解决思路
CrashLoopBackOff	容器启动后崩溃（代码报错、OOM、参数错误）	kubectl logs <pod-name> -n <namespace> --previous 查看上一次崩溃日志
OOMKilled	容器内存超出 limits 限制	调大 resources.limits.memory，或排查代码内存泄漏
ImagePullBackOff	镜像名称写错、私有仓库认证失败、网络不通	检查镜像 Tag、imagePullSecrets、以及容器运行时状态
Pending	资源不足（CPU/内存不够）、PVC 无法挂载、节点亲和性不匹配	kubectl describe pod 查看是否有 FailedScheduling 事件
Terminating	容器无法正常退出（被 preStop 钩子或 finalizer 阻塞）	检查应用是否收到 SIGTERM 后未优雅退出，必要时加 --force --grace-period=0 强制删除


二、 Node 节点故障排障

当节点出现异常时，集群将无法在上面正常调度或运行 Pod：
bash
kubectl get nodes
kubectl describe node <node-name>
请谨慎使用此类代码。
• Node 状态为 NotReady：
	• 检查节点上的 kubelet 服务状态：systemctl status kubelet。
	• 查看 kubelet 日志定位原因：journalctl -u kubelet -f 或查看 /var/log/kubelet.log。
	• 检查节点底层资源（磁盘 DiskPressure、内存 MemoryPressure、PID 资源是否耗尽）。
• 控制面（Control Plane）组件异常：
	• 登录 Master/Control-plane 节点，检查核心静态 Pod 或服务日志：
		• /var/log/kube-apiserver.log
		• /var/log/kube-controller-manager.log
		• /var/log/kube-scheduler.log

三、 网络与 Service 访问故障

如果 Pod 运行正常，但外部或集群内无法访问 Service：
1. 检查端点（Endpoints）是否存在：bash
kubectl get endpoints <service-name> -n <namespace>
请谨慎使用此类代码。

	• 如果 ENDPOINTS 为空，说明 Service 的 selector 标签与 Pod 的 labels 没对上。
2. 检查集群内置 DNS 解析：
	• 进入一个临时测试 Pod：kubectl run -it --rm debug --image=busybox:1.28 -- sh

1. 
	• 测试 DNS：nslookup kubernetes.default 或 nslookup <service-name>.<namespace>.svc.cluster.local。
	• 排查 CoreDNS 自身 Pod 是否正常运行：kubectl get pods -n kube-system -l k8s-app=kube-dns。
2. 检查网络插件（CNI）与 kube-proxy：
	• 确认 CNI 插件（如 Calico, Flannel, Cilium）的 Pod 是否处于 Running 状态。
	• 检查 kube-proxy 日志是否有异常报错或 IPVS/Iptables 规则加载失败。

# Reference 
https://kubernetes.io/zh-cn/docs/tasks/debug/debug-cluster/
