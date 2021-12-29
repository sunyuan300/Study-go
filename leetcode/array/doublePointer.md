# 双指针

双指针法（快慢指针法）在数组和链表的操作中是非常常见的，很多考察`数组、链表、字符串`等操作的面试题，都使用双指针法。

**核心思想**：`将一个数组当作两个数组使用,使用两个指针分别指向`。
![](https://tva1.sinaimg.cn/large/008eGmZEly1gntrds6r59g30du09mnpd.gif)

[leetcode27](https://leetcode-cn.com/problems/remove-element/)

```go
// nums = [0,1,2,2,3,0,4,2] // fast指向
// nums = [0,1,2,2,3,0,4,2] // slow指向
func removeElement(nums []int, val int) int {
	var slow int
	for fast:=0;fast<len(nums);fast++{
		if nums[fast] != val {
			nums[slow] = nums[fast]
			slow++
		}
	}
	return slow
}
```