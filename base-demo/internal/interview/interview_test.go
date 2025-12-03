package interview

import (
	"fmt"
	"sync"
	"testing"
)

// 接受一个int 参数，1<n<8,输出左右括号()的合法排列组合数组
// 列子
// 输入: 3
// 输出: ["()()()", "()(())", "(())()", "(()())", "((()))"]
func PrintOne(n int) []string {
	if n < 1 || n > 8 {
		fmt.Println("n is out of range")
		return []string{}
	}

	result := []string{}
	for i := 0; i < n; i++ {
		result = append(result, "()")
	}
	for i := 0; i < n; i++ {
		result = append(result, "("+result[i]+")")
	}
	for i := 0; i < n; i++ {
		result = append(result, "("+result[i]+")")
	}
	for i := 0; i < n; i++ {
		result = append(result, "("+result[i]+")")
	}
	for i := 0; i < n; i++ {
		result = append(result, "("+result[i]+")")
	}
	return result
}

// 接收一个数组，找出数组中乘积最大的非空连续子数组（子数组最少包含一个数字）
// 列子
// 输入: [2,3,-2,4]
// 输出: 6
// 解释: 子数组 [2,3] 有最大乘积 6
func PrintTwo(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for i := 0; i < len(nums); i++ {
		product := 1
		for j := i; j < len(nums); j++ {
			product *= nums[j]
		}
	}
	return max
}

// 创建 n个协程，按顺序打印 1 到 n，使用
func PrintThree(n int) int {

	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			fmt.Println(i)
		}(i)
	}
	wg.Wait()
	return n
}

func TestInterview(t *testing.T) {
	printOne := PrintOne(3)
	fmt.Println(printOne)
	printTwo := PrintTwo([]int{2, 3, -2, 4})
	fmt.Println(printTwo)

	printThree := PrintThree(10)
	fmt.Println(printThree)
}
