package db

type Statement struct {
	ID         string `json:"id"`
	Title      string `json:"text"`
	UserId     string `json:"userId"`
	CategoryId string `json:"categoryId"`
}

type Category struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	UserId string `json:"userId"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"` //md5 hash
}
