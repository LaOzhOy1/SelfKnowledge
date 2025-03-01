package main

import "fmt"

//func main() {
//	array := []int{10, 20, 30, 40}
//	slice1 := make([]int, 6)
//	n1 := copy(array, slice1)
//	fmt.Println(n1,array)
//
//
//	slice := make([]byte, 3)
//	n := copy(slice, "abcdef")
//	fmt.Println(n,slice)
//}

// 如果用 range 的方式去遍历一个切片，拿到的 Value 其实是切片里面的值拷贝。
func main() {
	slice := []int{10, 20, 30, 40}
	for index, value := range slice {
		value++
		fmt.Printf("value = %d , value-addr = %x , slice-addr = %x\n", value, &value, &slice[index])
	}
	for index, value := range slice {
		fmt.Printf("value = %d , value-addr = %x , slice-addr = %x\n", value, &value, &slice[index])
	}
	for index, value := range slice {
		slice[index] = value + 1
		fmt.Printf("value = %d , value-addr = %x , slice-addr = %x\n", value, &value, &slice[index])
	}
	for index, value := range slice {
		fmt.Printf("value = %d , value-addr = %x , slice-addr = %x\n", value, &value, &slice[index])
	}
}

//func main() {
//	//   arrayA的指针指向改变会影响goroutine
//
//	arrayA := [2]int{100, 200}
//	go testArrayPoint1(&arrayA) // 1.传数组指针
//	time.Sleep(time.Second)
//	arrayA = [2]int{0, 200}
//	time.Sleep(time.Second)
//	/*
//		这里是用字面量创建的一个 len = 2，cap = 2 的切片，这时候数组里面每个元素的值都初始化完成了。需要注意的是 [ ] 里面不要写数组的容量，因为如果写了个数以后就是数组了，而不是切片了。
//		arrayA := []int{100, 200}
//	*/
//	arrayB := arrayA[1:2:2]
//	testArrayPoint2(&arrayB) // 2.传切片
//
//
//	fmt.Printf("arrayA : %p , %v\n", &arrayA, arrayA)
//}

//func testArrayPoint1(x *[2]int) {
//	for true {
//		fmt.Printf("func Array : %p , %v\n", x, *x)
//		time.Sleep(time.Second/10)
//	}
//	//(*x)[1] += 100
//}
//
//func testArrayPoint2(x *[]int) {
//	fmt.Printf("func Array : %p , %v\n", x, *x)
//	(*x)[1] += 100
//}
//func main() {
//	//自动推导类型，同时进行初始化
//	s1 := []int{1, 2, 3, 4}
//	fmt.Println("s1=", s1)
//
//	//借助make的方式创建切片（类型 长度 容量）
//	s2 := make([]int, 5, 10)
//	fmt.Println(s2)
//	fmt.Printf("len=%d,cap=%d\n", len(s2), cap(s2))
//
//	//如果没有指定容量  容量和长度一样
//	s3 := make([]int, 5)
//	fmt.Printf("len=%d,cap=%d\n", len(s3), cap(s3))
//
//	s3 = s3[1:4]
//	fmt.Printf("len=%d,cap=%d\n", len(s3), cap(s3))
//	//fmt.Printf("invalid value = %d\n",s3[3])
//	s3 = append(s3,100)
//	fmt.Printf("len=%d,cap=%d\n", len(s3), cap(s3))
//
//	s3 = append(s3,200)
//	fmt.Printf("len=%d,cap=%d\n", len(s3), cap(s3))
//}
//
//
//
//func ArrayRemoveRepeated(arr []string) []string {
//	sort.Strings(arr)
//	i:= 0
//	var j int
//	for{
//		if i >= len(arr)-1 {
//			break
//		}
//		for j = i + 1; j < len(arr) && arr[i] == arr[j]; j++ {
//		}
//		arr= append(arr[:i+1], arr[j:]...)
//		i++
//	}
//	return  arr
//}
