[206. 反转链表](https://leetcode-cn.com/problems/reverse-linked-list/)
![](https://tva1.sinaimg.cn/large/008eGmZEly1gnrf1oboupg30gy0c44qp.gif)

首先定义一个cur指针，指向头结点，再定义一个pre指针，初始化为null。

反转：
1. 记录cur指向的下一个节点`temp=cur.Next`
2. 反转`cur.Next=pre`
3. 移动pre`pre=cur`
4. 移动cur`cur=temp`

```go
func reverseList(head *ListNode) *ListNode {
	var pre,cur,temp *ListNode
	cur = head
	for cur!=nil {
		temp = cur.Next
		cur.Next = pre
		pre = cur
		cur = temp
	}
	return pre

}
```

## 递归法
```go
func reverseList(head *ListNode) *ListNode {
	return help(nil,head)
}

func help(pre,cur *ListNode) *ListNode {
	if cur == nil {
		return pre
	}

	next := cur.Next
	cur.Next = pre
	return help(cur,next)
}
```