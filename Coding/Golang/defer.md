# Defer





defer关键字的实现跟go关键字很类似，不同的是它调用的是runtime.deferproc而不是runtime.newproc。

**在defer出现的地方，插入了指令call runtime.deferproc，然后在函数返回之前的地方，插入指令call runtime.deferreturn**

普通的函数返回时，汇编代码类似：

```
add xx SP
return
```

如果其中包含了defer语句，则汇编代码是：

```
call runtime.deferreturn，
add xx SP
return
```

goroutine的控制结构中，有一张表记录defer，调用runtime.deferproc时会将需要defer的表达式记录在表中，而在调用runtime.deferreturn的时候，则会依次从defer表中出栈并执行。

​	