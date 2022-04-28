package utils

/*
创建人员：云深不知处
创建时间：2022/4/27
程序功能：删除切片指定元素
*/

// DeleteStrSliceElms 删除切片指定元素（不改原切片）
func DeleteStrSliceElms(sl []string, elms ...string) []string {
	if len(sl) == 0 || len(elms) == 0 {
		return sl
	}
	// 先将元素转为 set
	m := make(map[string]struct{})
	for _, v := range elms {
		m[v] = struct{}{}
	}
	// 过滤掉指定元素
	res := make([]string, 0, len(sl))
	for _, v := range sl {
		if _, ok := m[v]; !ok {
			res = append(res, v)
		}
	}
	return res
}
