package main

import (
	"fmt"
	"reflect"
)

/*
*
printf("     对数组中指定位置上置位和判断该位\n");

	printf("--- by MoreWindows( http://blog.csdn.net/MoreWindows )  ---\n\n");
	//在数组中在指定的位置上写1
	int b[5] = {0};
	int i;
	//在第i个位置上写1
		// 原理是使用整数中的32位bit
	for (i = 0; i < 40; i += 3)
		b[i / 32] |= (1 << (i % 32));
	//输出整个bitset
	for (i = 0; i < 40; i++)
	{
		if ((b[i / 32] >> (i % 32)) & 1)
			putchar('1');
		else
			putchar('0');
	}
	putchar('\n');
	return 0;

————————————————
版权声明：本文为CSDN博主「MoreWindows」的原创文章，遵循CC 4.0 BY-SA版权协议，转载请附上原文出处链接及本声明。
原文链接：https://blog.csdn.net/morewindows/article/details/7354571
*/
func main() {
	var num = 1 << 16

	fmt.Println(num)
	fmt.Println(^num + 1)
	if num>>31 == 0 {
		fmt.Println(num)
	} else {
		fmt.Println(^num + 1)
	}
	fmt.Println(ReverseBitsInWord03(num))
	fmt.Println(reverseInt(num))
	fmt.Printf("%v", reflect.TypeOf(num))
	fmt.Printf("%b\n", num)
	fmt.Printf("%d\n", num)
	fmt.Printf("%x\n", num)
	fmt.Printf("%f\n", num)
}

// 转二进制
func ReverseBitsInWord03(Num int) int {
	Num = (Num&0x55555555)<<1 | (Num&0xAAAAAAAA)>>1

	Num = (Num&0x33333333)<<2 | (Num&0xCCCCCCCC)>>2

	Num = (Num&0x0F0F0F0F)<<4 | (Num&0xF0F0F0F0)>>4

	Num = (Num&0x00FF00FF)<<8 | (Num&0xFF00FF00)>>8

	Num = (Num&0x0000FFFF)<<16 | (Num&0xFFFF0000)>>16

	return Num
}

func reverseInt(num int) int {

	pow := 0

	for num > 0 {
		a := num % 10
		num /= 10
		tmp := pow*10 + a
		if tmp/10 != pow {
			return 0
		}
		pow = tmp
	}
	return pow
}
