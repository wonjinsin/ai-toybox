package domain

type Category struct {
	ID   int
	Name string
}

// AllowedCategories is the fixed set the AI classifies merchants into.
var AllowedCategories = []string{
	"식비", "카페/간식", "교통", "주거/통신", "생활/마트", "쇼핑",
	"의료/건강", "문화/여가", "여행", "교육", "경조/이체", "금융/수수료",
	"급여/수입", "기타",
}
