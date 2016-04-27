package GodoersToDo

import (
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"golang.org/x/crypto/bcrypt" //password hashing
)


//globals
var tpl *template.Template

func init() {
	tpl = template.Must(tpl.ParseGlob("template/*.html"))

	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/", index)
	r.HandleFunc("/login", login)
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
	//get session from memcache -> session.go
	_, session_id, err := getSession(request)

	//found a session redirect to dashboard(Problem: getting the session info again in dashboard)
	if err == nil {
		http.Redirect(response, request, `/dashboard?id=`+session_id, http.StatusSeeOther)
	}
	//else stay on index
	tpl.ExecuteTemplate(response, "index.html", nil)
}



//login process
func login(response http.ResponseWriter, request *http.Request){
	var session Session
	var user User
	ctx := appengine.NewContext(request)

	if request.Method == "POST" {
		email := request.FormValue("email")
		password := request.FormValue("password")

		//get the user with given email in datastore
		key := datastore.NewKey(ctx, "Users", email, 0, nil)
		err := datastore.Get(ctx, key, &user) //store info of User in datastore to user
		
		//login failed
		//wrong password || user email not found
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			log.Infof(ctx, "*** Error Info: Login Failed, given credentials not found in datastore. ***")
			session.Message = "Logged in Failed! \n Email or password incorrect"
		} else{
			//login success
			//create a new session for the user
			session_id := createSession(response, request, user)
			http.Redirect(response, request, `/dashboard?id=`+session_id, http.StatusSeeOther)
		}
	}
	tpl.ExecuteTemplate(response, "index.html", session)
}



func logout(response http.ResponseWriter, request *http.Request){

}



func dashboard(response http.ResponseWriter, request *http.Request){
	//get session from memcache -> session.go
	_, _, err := getSession(request)
	//no session found anywhere(means not login)
	if err != nil {
		http.Redirect(response, request, `/`, http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(response, "index.html", session)
}



func register(response http.ResponseWriter, request *http.Request){
	var session Session
	var user User
	ctx := appengine.NewContext(request)

	if request.Method == "POST" {
		firstname := request.FormValue("firstname")
		lastname  := request.FormValue("lastname")
		email     := request.FormValue("email")

		password1 := request.FormValue("password1")
		password2 := request.FormValue("password2")

		user.Email = email
		key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
		err := datastore.Get(ctx, key, &user)
		
		//if there is no errors in getting the email in datastore, it means that 
		//the email is already taken and therefore not unique
		if err == nil{
			log.Infof(ctx, "*** Error Info: In register, email not unique ***")
			session.Message = "Email already exists \n "
		}
		//password confirmations not match
		if password1 != password2 {
			log.Infof(ctx, "*** Error Info: In register, password confirmations not match ***")
			session.Message += "Password Confirmation Not Match!"
			tpl.ExecuteTemplate(response, "register.html", session)
			return
		}
		//no errors
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In register, password hashing: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

		newUser := User{
			FirstName: firstname,
			LastName:  lastname,
			Email:     email,
			Password:  string(hashed_password),
		}
		key = datastore.NewKey(ctx, "Users", newUser.Email, 0, nil)
		key, err = datastore.Put(ctx, key, &newUser)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In register, failed to save newUser to datastore: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

		session_id := createSession(response, request, newUser)
		http.Redirect(response, request, "/dashboard?id="+session_id, http.StatusSeeOther)
		
	}
	tpl.ExecuteTemplate(response, "register.html", session)
}


//go get github.com/gorilla/mux

//the session_id == uuid == cookie.Value is being passed in the url 