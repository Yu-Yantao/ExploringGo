package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func MergeMeetingRooms(input string) string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	var intervals [][]int

	// 1. 解析输入到二维切片中
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		start, _ := strconv.Atoi(fields[0])
		end, _ := strconv.Atoi(fields[1])
		intervals = append(intervals, []int{start, end})
	}

	if len(intervals) == 0 {
		return ""
	}

	// 2. 核心：按开始时间升序排序（确保后面的时间只能跟当前结尾比）
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	// 3. 遍历合并
	var merged [][]int
	merged = append(merged, intervals[0])

	for i := 1; i < len(intervals); i++ {
		curr := intervals[i]
		lastMerged := merged[len(merged)-1]

		// 如果当前会议的开始时间 <= 上一个会议的结束时间，说明相接或重叠，进行合并
		if curr[0] <= lastMerged[1] {
			if curr[1] > lastMerged[1] {
				lastMerged[1] = curr[1] // 延长当前会议的结束时间
			}
		} else {
			// 如果当前开始时间 > 上一个结束时间（比如 60 > 45），说明中间有空闲，不合并！
			// 直接作为独立的段放入结果集
			merged = append(merged, curr)
		}
	}
	// 4. 统计总分钟数并格式化输出
	var sb strings.Builder
	totalMinutes := 0

	for i, interval := range merged {
		totalMinutes += (interval[1] - interval[0])
		sb.WriteString(fmt.Sprintf("[%d,%d]", interval[0], interval[1]))
		if i < len(merged)-1 {
			sb.WriteString(" ")
		}
	}

	return fmt.Sprintf("%s | %d", sb.String(), totalMinutes)
}

func main() {
	// 样例 1 测试：45到60没会议，断开不合并
	input1 := `
0 30
30 45
60 75
	`
	fmt.Println(MergeMeetingRooms(input1)) // 输出: [0,45] [60,75] | 60
}
