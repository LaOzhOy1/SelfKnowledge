# FMT



## fmt.Println

- 内部调用printArg方法，默认传入 verb 参数 ’v‘
  ![image-20220627002643761](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270125927.png)
- 处理非基础类型参数时，以他的ValueOf值作为输出
  ![image-20220627002615167](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206270125776.png)

- printValue 根据Kind 类型进行不同的输出
  - Channel 输出为其指针
  - Map、Fun、Slice 、Struct 、Interface、Array、Slice输出为结构化字符串
  - Ptr 额外输出&并再一次调用printValue