package main

import (
	"fmt"
)

// 定义一个Robot结构体
type Robot struct {
	Name  string
	Color string
	Model string
}

func main() {
	fmt.Println("深拷贝 内容一样，改变其中一个对象的值时，另一个不会变化。")
	robot1 := Robot{
		Name:  "小白-X型-V1.0",
		Color: "白色",
		Model: "小型",
	}

	// 值拷贝
	robot2 := robot1
	fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, &robot1)
	fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, &robot2)

	fmt.Println("修改Robot1的Name属性值")

	robot1.Name = "小白-X型-V1.1"

	fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, &robot1)
	fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, &robot2)

	/**
	表达式new(T)将创建一个T类型的匿名变量，所做的是为T类型的新值分配并清零一块内存空间，

	然后将这块内存空间的地址作为结果返回，而这个结果就是指向这个新的T类型值的指针值，返回的指针类型为*T。

	new创建的内存空间位于heap上，空间的默认值为数据类型默认值。如：new(int) 则 *p为0，new(bool) 则 *p为false。
	*/
	// new 返回指针
	robotNew := new(Robot)
	robotNew.Color = "白色"

	robotCopy := robotNew
	robotCopy.Color = "hongse"
	fmt.Print(robotNew.Color)
}

/**

package main

import (
   "fmt"
)

// 定义一个Robot结构体
type Robot struct {
   Name  string
   Color string
   Model string
}

func main() {

   fmt.Println("浅拷贝 内容和内存地址一样，改变其中一个对象的值时，另一个同时变化。")
   robot1 := Robot{
      Name:  "小白-X型-V1.0",
      Color: "白色",
      Model: "小型",
   }
   robot2 := &robot1
   fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, &robot1)
   fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, robot2)

   fmt.Println("在这里面修改Robot1的Name和Color属性")
   robot1.Name = "小黑-X型-V1.1"
   robot1.Color = "黑色"

   fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, &robot1)
   fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, robot2)

}



fmt.Println("浅拷贝 使用new方式")
   robot1 := new(Robot)
   robot1.Name = "小白-X型-V1.0"
   robot1.Color = "白色"
   robot1.Model = "小型"

   robot2 := robot1
   fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, robot1)
   fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, robot2)

   fmt.Println("在这里面修改Robot1的Name和Color属性")
   robot1.Name = "小蓝-X型-V1.2"
   robot1.Color = "蓝色"

   fmt.Printf("Robot 1：%s\t内存地址：%p \n", robot1, robot1)
   fmt.Printf("Robot 2：%s\t内存地址：%p \n", robot2, robot2)


*/
