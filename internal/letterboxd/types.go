package letterboxd

const BaseURL = "https://letterboxd.com"

type DiaryEntry struct {
	Date    string
	Title   string
	FilmURL string
	Rating  string
	Rewatch bool
	Review  bool
}

type Profile struct {
	Stats     []ProfileStat
	Favorites []FavoriteFilm
	Recent    []ProfileRecent
}

type ProfileStat struct {
	Label string
	Value string
	URL   string
}

type FavoriteFilm struct {
	Title   string
	FilmURL string
	Year    string
}

type ProfileRecent struct {
	Summary string
	FilmURL string
}

type WatchlistItem struct {
	Title   string
	FilmURL string
	Year    string
}

type Film struct {
	Title       string
	Year        string
	Description string
	Director    string
	AvgRating   string
	Runtime     string
	URL         string
	Slug        string
	FilmID      string
	WatchlistID string
	InWatchlist bool
	WatchlistOK bool
	ViewingUID  string
	Cast        []string
	UserRating  string
	UserStatus  string
}

type ActivityItem struct {
	ID       string
	Summary  string
	When     string
	Title    string
	FilmURL  string
	Rating   string
	Kind     string
	Actor    string
	ActorURL string
	Parts    []SummaryPart
}

type SummaryPart struct {
	Text string
	Kind string
}

type SearchResult struct {
	Title   string
	Year    string
	FilmURL string
	Slug    string
	FilmID  string
}
