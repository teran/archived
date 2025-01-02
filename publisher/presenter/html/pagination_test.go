package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPagination(t *testing.T) {
	r := require.New(t)

	pagination := NewPagination(305, 10, 3, "/some/page")
	r.Equal(Pagination{
		HasPrevious:     true,
		PreviousPageNum: 9,
		PreviousPageURL: "/some/page?page=9",
		Pages: []Page{
			{IsCurrent: true, URL: "/some/page?page=10", Number: 10},
			{IsCurrent: false, URL: "/some/page?page=11", Number: 11},
			{IsCurrent: false, URL: "/some/page?page=12", Number: 12},
			{IsCurrent: false, URL: "/some/page?page=13", Number: 13},
		},
		HasNext:     true,
		NextPageNum: 14,
		NextPageURL: "/some/page?page=14",
	}, pagination)
}
