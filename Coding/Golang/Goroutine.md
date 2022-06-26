# 状态

- Runable
- Runging
- SysCall
- Waiting
- Gdead
- GcopyStack
- _Gpreempted // 指的是当前协程因为SuspendG的原因需要暂停自己的活动，而且要等待SuspendG 将自己的Status装为 _Gwaiting 才能被Ready 方法唤醒
- Gscan