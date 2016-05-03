package GodoersToDo

import (
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"golang.org/x/crypto/bcrypt" //password hashing
	"google.golang.org/appengine/memcache"
	"encoding/json"
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
	r.HandleFunc("/api/passw_check", passw_check)

	r.Handle("/favicon.ico", http.NotFoundHandler())

	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("js"))))
	http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.Dir("img"))))
}



func index(response http.ResponseWriter, request *http.Request) {
	//get session from memcache -> session.go
	var session Session
	_, session_id, err := getSession(request)
	session.Session_id = session_id

	//found a session redirect to dashboard(Problem: getting the session info again in dashboard)
	if err == nil {
		http.Redirect(response, request, `/dashboard?id=`+session.Session_id, http.StatusSeeOther)
	}
	//else stay on index
	tpl.ExecuteTemplate(response, "index.html", nil)
}


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
		//wrong password || no user in datastore
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			log.Infof(ctx, "*** Error Info: Login Failed, given credentials not found in datastore. ***")
			session.Message = "Logged in Failed! \n Email or password incorrect"
		} else{
			//login success
			//create a new session for the user
			session.Session_id = createSession(response, request, user)
			http.Redirect(response, request, `/dashboard?id=`+session.Session_id, http.StatusSeeOther)
		}
	}
	tpl.ExecuteTemplate(response, "index.html", session)
}


func logout(response http.ResponseWriter, request *http.Request){
	//delete cookie and item in memcache effectively destroying the session
	deleteSession(response, request)
	//after the cookie is deleted and the session in memcache redirect to index
	//and the user will not be able to go back to dashbaord because the user
	//has no session set.
	http.Redirect(response, request, `/`, http.StatusSeeOther)
}


func dashboard(response http.ResponseWriter, request *http.Request){
	ctx := appengine.NewContext(request)
	var session Session
	var user User
	//get session from memcache -> session.go
	_, session_id, err := getSession(request)
	//no session found anywhere(means not login)
	if err != nil {
		//redirect to index
		http.Redirect(response, request, `/`, http.StatusSeeOther)
		return //is this needed?
	}

	session.Session_id = session_id
	//retrieve session in memcache
	item, err := memcache.Get(ctx, session_id)

	//if no session was found in memcache then invoke logout that
	//effectively deletes the session.
	//this is a guard for when a cookie is not found, logout which calls deleteSession
	//will use the url value when there is no cookie and will use that url value to
	//reference the item in memcache to delete it.
	if err != nil{
		logout(response, request)
		return //probably dont need this
	}

	//found a session, then unmarshal the user
	json.Unmarshal(item.Value, &user)
	session.User = user
	//pass session which has the user information to dash.html 
	tpl.ExecuteTemplate(response, "dash.html", session)
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
		//generate a key and check the datastore to check if the user email
		//already exists in datastore
		key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
		err := datastore.Get(ctx, key, &user)
		
		//if there is no errors in getting the email in datastore, it means that 
		//the email is already taken and therefore not unique
		if err == nil{
			log.Infof(ctx, "*** Error Info: In register, email not unique ***")
			//if the user email is already in datastore then generate an error message 
			//and pass it to register.html to show to the user.
			session.Message = "Email already exists \n "
			tpl.ExecuteTemplate(response, "register.html", session)
			return
		}
		//password confirmations not match error
		if password1 != password2 {
			log.Infof(ctx, "*** Error Info: In register, password confirmations not match ***")
			//generate error message
			session.Message += "Password Confirmation Not Match!"
			//if the password confirmation fails then generate an error message 
			//and pass it to register.html to show to the user.
			tpl.ExecuteTemplate(response, "register.html", session)
			return
		}

		//no errors in the user inputs
		//secure the password using bcrypt and create the new user with the given information from
		//post
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
		if err != nil {
			//server error
			log.Errorf(ctx, "*** Error Debug: In register, password hashing: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}
		
		//create new user with given values
		newUser := User{
			FirstName: firstname,
			LastName:  lastname,
			Email:     email,
			Password:  string(hashed_password),
		}

		//generate new key to use for saving the user to datastore
		key = datastore.NewKey(ctx, "Users", newUser.Email, 0, nil)
		key, err = datastore.Put(ctx, key, &newUser) //save user to datastore
		if err != nil {
			//server error
			log.Errorf(ctx, "*** Error Debug: In register, failed to save newUser to datastore: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

		//create a session for the new user so the user will be automatically logged in after 
		//registration
		session_id := createSession(response, request, newUser)
		http.Redirect(response, request, "/dashboard?id="+session_id, http.StatusSeeOther)
		
	}
	tpl.ExecuteTemplate(response, "register.html", session)
}

func profile(response http.ResponseWriter, request *http.Request){
	var session Session
	var user User
	ctx := appengine.NewContext(request)
	item, session_id, err := getSession(request)
	json.Unmarshal(item.Value, &user)
	session.User = user


	if request.Method == "POST" {
		firstname := request.FormValue("firstname")
		lastname  := request.FormValue("lastname")
		email     := request.FormValue("email")

		password1 := request.FormValue("password1")
		password2 := request.FormValue("password2")

		if err != nil {
			key := datastore.NewKey(ctx, "User", email, 0, nil)
			err := datastore.Get(ctx, key, &user)
			if err != nil{
				log.Errorf(ctx, "*** Error Debug: In Profile, User not Found: %v ***", err)
				logout(response, request)
			}

		}else{
			json.Unmarshal(item.Value, &user)
		}

		if user.Email != email {
			var checkuser User
			key := datastore.NewKey(ctx, "Users", user.Email, 0, nil)
			err := datastore.Get(ctx, key, &checkuser)
			
			//if there is no errors in getting the email in datastore, it means that 
			//the email is already taken and therefore not unique
			if err == nil{
				log.Infof(ctx, "*** Error Info: In profile, email not unique ***")
				//if the user email is already in datastore then generate an error message 
				//and pass it to register.html to show to the user.
				session.Message = "Email already exists \n "
				tpl.ExecuteTemplate(response, "profile.html", session)
				return
			}
		}

		//password confirmations not match error
		if password1 != password2 {
			log.Infof(ctx, "*** Error Info: In profile, password confirmations not match ***")
			//generate error message
			session.Message += "Password Confirmation Not Match!"
			//if the password confirmation fails then generate an error message 
			//and pass it to register.html to show to the user.
			tpl.ExecuteTemplate(response, "profile.html", session)
			return
		}
		oldEmail := user.Email
		user.Email = email
		user.FirstName = firstname
		user.LastName = lastname
		hashed_password, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
		if err != nil {
			//server error
			log.Errorf(ctx, "*** Error Debug: In profile, password hashing: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

		user.Password = string(hashed_password)
		json, err := json.Marshal(user)
		if err != nil {
			//error marshalling user
			log.Errorf(ctx, "*** Error Debug: In Profile json.Marshal: %v ***", err)
			//http.Error() replies to the request with the specified error message and HTTP code. 
			//The error message should be plain text.
			http.Error(response, err.Error(), 500)
			return
		}
		//for debugging purposes: paste the cookie id from the terminal to memcache viewer
		//to see if the user(json) is being cached in memcache
		//log.Infof(ctx, "Cookie Id:" + " " + cookie.Value)
		log.Infof(ctx, session_id)
		m := memcache.Item{
			Key:   session_id,
			Value: json,
		}
		memcache.Set(ctx, &m)

		key := datastore.NewKey(ctx, "Users", oldEmail, 0, nil)
		key, err = datastore.Put(ctx, key, &user) //save user to datastore
		if err != nil {
			//server error
			log.Errorf(ctx, "*** Error Debug: In register, failed to save newUser to datastore: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

	}

	tpl.ExecuteTemplate(response, "profile.html", session)
}

//go get github.com/gorilla/mux

//the session_id == uuid == cookie.Value is being passed in the url res]