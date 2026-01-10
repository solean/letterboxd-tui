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
	Recent    []string
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
	Cast        []string
	UserRating  string
	UserStatus  string
}

type ActivityItem struct {
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
