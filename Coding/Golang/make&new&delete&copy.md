# Make 和 New 的区别



## Make

func make(t Type, size ...IntegerType) Type

- make 只能用于内置数据结构
- make 返回的是对应的类型的引用
- 对于Slice，make([]int, 0, 10)相当于在底层创建了长度为10的数组，并返回长度为0的切片
- 对于map，不用去声明map所对应的长度，如果长度小于创建map的默认大小会被忽略
- 对于Channel， 0和非0 差别在于缓冲区





## New

func new(Type) *Type

- 开辟对于数据类型的内存空间并返回指针，指针指向类型的默认空值

- 对于通过new初始化内置数据结构

  - Slice，会开辟一个底层数组长度为0的内存空间，返回*[]int，无法直接使用
  - Map，会开辟一个底层Map长度为0的内存空间，返回*map，无法直接使用[][]

  - Channe 会申请一个内存空间地址, 相当于Make（Channel， 0），可以直接使用





![image-20220626035432686](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206260354703.png)







## Delete

删除Map数组中特殊的key下的键值对





## Copy

对数据进行复制，返回同步的元素数量，是src 和 dst长度的最小值