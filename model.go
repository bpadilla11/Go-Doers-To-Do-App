package GodoersToDo


const (
	todo = "ToDo"
	in_progress = "InProgress"
	done = "Done"
)


type User struct {
	FirstName string
	LastName  string
	Email     string //primary key(unique)
	Password  string
}


type Session struct {
	User  User
	State bool
}


type ToDo struct {
	User  User
	Name  string
	Task  []Task
}


type Task struct {
	User        User
	Description string
	Status      string  //choices = todo | in_progress | done
	Date        string  //t := time.Now()
					    //fmt.Println(t.Format("03-02-2006"))
}