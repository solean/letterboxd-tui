package letterboxd

import "fmt"

type DiarySort string

const (
	DiarySortDefault       DiarySort = ""
	DiarySortAddedEarliest DiarySort = "added-earliest"
	DiarySortRating        DiarySort = "entry-rating"
)

type WatchlistSort string

const (
	WatchlistSortDefault      WatchlistSort = ""
	WatchlistSortName         WatchlistSort = "name"
	WatchlistSortDateEarliest WatchlistSort = "date-earliest"
	WatchlistSortRelease      WatchlistSort = "release"
	WatchlistSortRating       WatchlistSort = "rating"
)

func diaryURL(username string, page int, sort DiarySort) string {
	if sort != "" {
		if page > 1 {
			return fmt.Sprintf("%s/%s/diary/films/by/%s/page/%d/", BaseURL, username, sort, page)
		}
		return fmt.Sprintf("%s/%s/diary/films/by/%s/", BaseURL, username, sort)
	}
	if page > 1 {
		return fmt.Sprintf("%s/%s/diary/films/page/%d/", BaseURL, username, page)
	}
	return fmt.Sprintf("%s/%s/diary/", BaseURL, username)
}

func watchlistURL(username string, page int, sort WatchlistSort) string {
	if sort != "" {
		if page > 1 {
			return fmt.Sprintf("%s/%s/watchlist/by/%s/page/%d/", BaseURL, username, sort, page)
		}
		return fmt.Sprintf("%s/%s/watchlist/by/%s/", BaseURL, username, sort)
	}
	if page > 1 {
		return fmt.Sprintf("%s/%s/watchlist/page/%d/", BaseURL, username, page)
	}
	return fmt.Sprintf("%s/%s/watchlist/", BaseURL, username)
}
