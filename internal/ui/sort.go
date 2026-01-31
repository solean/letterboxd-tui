package ui

import "github.com/solean/letterboxd-tui/internal/letterboxd"

type diarySort int

const (
	diarySortRecent diarySort = iota
	diarySortOldest
	diarySortRating
	diarySortCount
)

type watchlistSort int

const (
	watchlistSortAdded watchlistSort = iota
	watchlistSortTitle
	watchlistSortRating
	watchlistSortYearNewest
	watchlistSortCount
)

func (s diarySort) next() diarySort {
	if s+1 >= diarySortCount {
		return diarySortRecent
	}
	return s + 1
}

func (s watchlistSort) next() watchlistSort {
	if s+1 >= watchlistSortCount {
		return watchlistSortAdded
	}
	return s + 1
}

func (m Model) diarySortLabel() string {
	switch m.diarySort {
	case diarySortOldest:
		return "Oldest"
	case diarySortRating:
		return "Rating"
	default:
		return "Recent"
	}
}

func (m Model) watchlistSortLabel() string {
	switch m.watchlistSort {
	case watchlistSortTitle:
		return "Title"
	case watchlistSortRating:
		return "Rating (avg)"
	case watchlistSortYearNewest:
		return "Year (newest)"
	default:
		return "Added"
	}
}

func (m Model) diarySortParam() letterboxd.DiarySort {
	switch m.diarySort {
	case diarySortOldest:
		return letterboxd.DiarySortAddedEarliest
	case diarySortRating:
		return letterboxd.DiarySortRating
	default:
		return letterboxd.DiarySortDefault
	}
}

func (m Model) watchlistSortParam() letterboxd.WatchlistSort {
	switch m.watchlistSort {
	case watchlistSortTitle:
		return letterboxd.WatchlistSortName
	case watchlistSortRating:
		return letterboxd.WatchlistSortRating
	case watchlistSortYearNewest:
		return letterboxd.WatchlistSortRelease
	default:
		return letterboxd.WatchlistSortDefault
	}
}
