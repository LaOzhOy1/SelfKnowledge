package main

import "fmt"

/*
*
小明目前在做一份毕业旅行的规划。打算从北京出发，
分别去若干个城市，然后再回到北京，
每个城市之间均乘坐高铁，且每个城市只去一次。
由于经费有限，希望能够通过合理的路线安排尽可能的省一些路上的花销。
给定一组城市和每对城市之间的火车票的价钱，找到每个城市只访问一次并返回起点的最小车费花销。
*/
func main() {

	var cityNum int
	_, _ = fmt.Scanln(&cityNum)

	var arr [][]int
	for i := 0; i < cityNum; i++ {
		ar := make([]int, cityNum)
		for j := 0; j < cityNum; j++ {
			fmt.Scan(&ar[j])
		}
		arr = append(arr, ar)
	}

	V := 1 << (cityNum - 1)
	var dp [][]int
	for i := 0; i < cityNum; i++ {
		// 子集 0000
		ar := make([]int, V)
		dp = append(dp, ar)
		dp[i][0] = 0
	}

	for j := 1; j < V; j++ {
		for i := 0; i < cityNum; i++ {
			dp[i][j] = 1<<16 - 1
			//并位与上1，等于0，说明是从i号城市出发，经过城市子集V[j]，回到起点0号城市
			if i == 0 || j>>(i-1)&1 == 0 {
				// 遍历剩下的城市 10000000000000000000000000000000
				for k := 1; k < cityNum; k++ {
					if j>>(k-1)&1 == 1 {
						dp[i][j] = min(dp[i][j], arr[i][k]+dp[k][j^(1<<(k-1))])
					}
				}
			}
		}
	}
	fmt.Println(dp[0][V-1])
}
func min(a int, b int) int {
	if a > b {
		return b
	} else {
		return a
	}

}
