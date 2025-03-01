# Docker


## Docker的核心原理

Docker 通过namespace实现资源隔离，通过cgroup实现资源限制，使用Copy-On-Write实现文件高效操作


### namespace的6项隔离

- 文件系统
- 网络
- 独立的主机名称
- 进程
- 进程间通信
- 权限
