# quickstartGo

#### 介绍
Go 学习总结


## 主要内容

- ## Go学习--基础篇
    1. ### 数据结构(dataStruct 文件夹中)  
        - struct
        - map
        - slice
        - tree
        - heap
        - stack
        - queue
            - ArrayList
            - LinkedList
        - set
            - HashSet

    2. ### 语法 （basic.grammar 文件夹中）
        - 变量与拷贝（copy）
        - struct比较与==(equals)
        - tag的使用(tag)
        - defer(defer)
        - iota(enum)
        - recover（Exception）
        - rune（rune）
        - 函数调用（invoker）
        - io输入输出流（io）
    3. ### 协程
        -- 协程的同步
    4. ### Channel 
        -- 死锁的情况
        -- 使用Channel遍历二叉树
        -- select的使用
        -- waitGroup的使用
- ## Go学习--进阶篇
- ## Go学习--项目实践
    1. ### 单任务爬虫（无工作队列）
        - 工程总体流程（engine/engine）
        - 工程所需结构体与接口定义（engine/types）
        - 抓取器（fetcher）
        - 解析器（parser）
    2. ### 并发爬虫（单工作队列）
        - 工程总体流程（engine/concurrentEngine）
        - 工程所需结构体与接口定义（engine/types）
        - 抓取器（fetcher.go）
        - 解析器（parser）
            - 城市解析器（citylist.go）
            - 人物列表解析器（personList.go）
            - 人物解析器（personProfile.go）
        - 存储（saver/itemSaver.go）
        - 调度器（SimpleScheduler.go）
        - 工作者（SimpleWorker.go）
    3. ### 并发爬虫（多工作队列）
        - 工程总体流程（engine/concurrentEngineWithQueue.go）
        - 工程所需结构体与接口定义（engine/types）
        - 抓取器（fetcher.go）
        - 解析器（parser）
            - 城市解析器（citylist.go）
            - 人物列表解析器（personList.go）
            - 人物解析器（personProfile.go）
        - 存储（saver/itemSaver.go）
        - 调度器（QueueScheduler.go.go.go）
        - 工作者（QueueWorker.go.go）
   4. ### 分布式爬虫（To do）
   5. ### GRPC入门（grpc包）
        - cmd包
            存放wire 与 wire_gen 文件之外
            存放cobra编译连接后的可执行文件命令
            - client 进行tcp链接，生成装配客户端请求，请求服务端
            - grpc 开启grpc服务，监听Ctrl+C事件，优雅关机
        - pkg包
            存放自定义工具包如优雅停机实现
        - config包
            存放用于wire生成对象时所需的配置对象
        - server包
        存放grpc server对象构造函数
        - controller包
            - grpc包
            存放protobuf生成的server对象（可以内含usecase/service），调用usecase（service）
        - proto
            存放proto文件与proto文件生成的内容，包括：
            - 生成调用的客户端代码
            - 生成服务端服务接口
            - 生成服务调用过程所需的对象



- ## Go学习--leecode题解（leecode）
    1. ### BFS
        1.迷宫最短路径