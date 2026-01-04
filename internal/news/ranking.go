package news

import "sort"

func RankByScore(a []Article) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].RelevanceScore > a[j].RelevanceScore
	})
}
