package GodoersToDo

type User struct {
	Id        int64
	FirstName string
	LastName  string
	Email     string
	Password  string
}


type Session struct {
	User    User
	ToDos   []ToDo
	Message string
	Session_id string
}


type ToDo struct {
	ToDoId  int64
	UserId  int64
	Content string
	Date    string
	Photo_Link   string
	Photo_Media  string
}