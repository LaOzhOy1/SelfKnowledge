# Make 和 New 的区别



## Make

func make(t Type, size ...IntegerType) Type

- make 只能用于内置数据结构

- make 返回的是对应的类型的引用

- 对于Slice，make([]int, 0, 10)相当于在底层创建了长度为10的数组，并返回长度为0的切片
  ![image-20220626235729539](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206262357566.png)

- 对于Map，不用去声明map所对应的长度，如果长度小于创建map的默认大小会被忽略，map会在堆上分配内存,并返回他的对应的值引用，即使他的函数方法返回的是*map。

  【对于长度小于8的map调用】

  ![image-20220627022447647](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270224669.png)

  ![image-20220627022529595](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270225613.png)

  【对于长度大于8的map调用】

  ![image-20220627022647363](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270226382.png)

  

  ![image-20220627025045898](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270252167.png)

  ![image-20220627020507492](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270205527.png)



- 对于Channel， 0和非0 差别在于缓冲区



![image-20220623131510022](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270252704.png)



![image-20220627025748878](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270257919.png)



## New

func new(Type) *Type

- new返回指向新分配的该类型的地值，不是空指针（nil），new只分配内存，不初始化内存。**所谓的初始化就是给类型赋初值，比如字符为空，整型为0, 逻辑值为false等。**

- 对于通过new初始化内置数据结构

  - Slice，会开辟一个底层数组长度为0的内存空间，返回*[]int，无法直接使用
  - Map，会开辟一个底层Map长度为0的内存空间，返回*map，无法直接使用[][]

  - Channe 会申请一个内存空间地址,返回*chan，可以直接使用





![image-20220626035432686](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206260354703.png)







## Delete

删除Map数组中特殊的key下的键值对





## Copy

对数据进行复制，返回同步的元素数量，是src 和 dst长度的最小值