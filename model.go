package GodoersToDo

const (
	queued = "queued"
	done = "done"
)


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
	Files      []File
}


type ToDo struct {
	ToDoId  int64
	UserId  int64
	Content string
	Status  string
	Date    string
	Photo_Link   string
	Photo_Media  string
}


type File struct{
	Name string
	Source_Link string
	Download_Link string
}