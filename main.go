package GodoersToDo

import (
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"golang.org/x/crypto/bcrypt" //password hashing
	"io"
)


//globals
var tpl *template.Template

func init() {
	tpl = template.Must(tpl.ParseGlob("template/*.html"))

	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", index)
	r.HandleFunc("login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/dashboard", dashboard)
	r.HandleFunc("/register", register)

	//ajax requests
	r.HandleFunc("/api/email_check", email_check)

	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("js"))))
	http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.Dir("img"))))
}



func index(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		login(response, request)
	}
	tpl.ExecuteTemplate(response, "index.html", nil)
}



//login process
func login(response http.ResponseWriter, request *http.Request){
	email := request.FormValue("email")
	password := request.FormValue("password")

	ctx := appengine.NewContext(request)

	//get the user with given email in datastore
	key := datastore.NewKey(ctx, "Users", email, 0, nil)

	var user User
	err := datastore.Get(ctx, key, &user) //store info of User in datastore to user
	
	//login failed
	//wrong password || user email not found
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		var s Session
		s.State = false
		tpl.ExecuteTemplate(response, "index.html", s)
		return
	}

	//create session cookie and/or url?

	io.WriteString(response, user.Email)
	//http.Redirect(response, request, "/dashboard", 302) //302 http status found
	tpl.ExecuteTemplate(response, "index.html", nil)
}



func logout(response http.ResponseWriter, request *http.Request){

}



func dashboard(response http.ResponseWriter, request *http.Request){

}



func register(response http.ResponseWriter, request *http.Request){
	/*ctx := appengine.NewContext(request)
	
	hashed_password, err := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
	if err != nil {
		//error hashing password
		return
	}

	user := User{
		FirstName: "1",
		LastName:  "2",
		Email:     "1@2.com",
		Password: string(hashed_password),
	}

	key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
	key, err = datastore.Put(ctx, key, &user)

	if err != nil {
		//error in saving user info in datastore
		return 
	}

	//create session via cookie and/or url

	/*if request.Method == "POST" {
		firstname := request.FormValue("firstname")
		lastname  := request.FormValue("lastname")
		email     := request.FormValue("email")

		password1 := request.FormValue("password1")
		password2 := request.FormValue("password2")
		if password1 != password2 {
			//error
		}
	}*/
	//io.WriteString(response, user.Email)
	tpl.ExecuteTemplate(response, "index.html", nil)
}


//go get github.com/gorilla/mux