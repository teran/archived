package html

import "strconv"

type Pagination struct {
	HasPrevious     bool
	PreviousPageNum uint64
	PreviousPageURL string
	Pages           []Page
	HasNext         bool
	NextPageNum     uint64
	NextPageURL     string
}

type Page struct {
	IsCurrent bool
	Number    uint64
	URL       string
}

func NewPagination(total, current, maxPages uint64, pageURL string) Pagination {
	pages := []Page{}
	max := current + maxPages
	if current+maxPages > total {
		max = total
	}
	for i := current; i <= max; i++ {
		isCurrent := false
		if i == current {
			isCurrent = true
		}
		pages = append(pages, Page{
			IsCurrent: isCurrent,
			Number:    i,
			URL:       pageURL + "?page=" + strconv.FormatUint(i, 10),
		})
	}

	var (
		hasPrevious     bool
		previousPageNum uint64
		previousPageURL string
	)
	if current > 1 {
		hasPrevious = true
		previousPageNum = current - 1
		previousPageURL = pageURL + "?page=" + strconv.FormatUint(previousPageNum, 10)
	}

	var (
		hasNext     bool
		nextPageNum uint64
		nextPageURL string
	)
	if max < total {
		hasNext = true
		nextPageNum = current + maxPages + 1
		nextPageURL = pageURL + "?page=" + strconv.FormatUint(nextPageNum, 10)
	}

	return Pagination{
		HasPrevious:     hasPrevious,
		PreviousPageNum: previousPageNum,
		PreviousPageURL: previousPageURL,
		Pages:           pages,
		HasNext:         hasNext,
		NextPageNum:     nextPageNum,
		NextPageURL:     nextPageURL,
	}
}
