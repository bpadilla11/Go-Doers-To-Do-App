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
	"google.golang.org/cloud/storage"
	"strconv"
	"strings"	
	"io"
)

const gcsBucket = "todolist-1292.appspot.com"

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
	r.HandleFunc("/profile", profile)
	r.HandleFunc("/files", files)

	//ajax requests
	r.HandleFunc("/api/email_check", email_check)
	r.HandleFunc("/api/passw_check", passw_check)
	r.HandleFunc("/todo", todo)
	r.HandleFunc("/todo?todo=", todo)
	r.HandleFunc("/api/filehelper", filehelper)

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
		//key := datastore.NewKey(ctx, "Users", email, 0, nil)
		//err := datastore.Get(ctx, key, &user) //store info of User in datastore to user
		q := datastore.NewQuery("Users").Filter("Email =", email).KeysOnly()
		i, _ := q.Count(ctx)

		keys, _ := q.GetAll(ctx, nil)
		if i > 0 {
			datastore.Get(ctx, keys[0], &user)
		}
		
		//login failed
		//wrong password || no user in datastore
		if i != 1 || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
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
	//ctx := appengine.NewContext(request)
	var session Session
	var user User
	//var todo ToDo
	
	item, session_id, err := getSession(request)
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
	session.Session_id = session_id

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
		q := datastore.NewQuery("Users").Filter("Email =", email)
		i, _ := q.Count(ctx)
		
		//if there is no errors in getting the email in datastore, it means that 
		//the email is already taken and therefore not unique
		if i != 0{
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
		key := datastore.NewIncompleteKey(ctx, "Users", nil)
		key, err = datastore.Put(ctx, key, &newUser) //save user to datastore
		newUser.Id = key.IntID()
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

	//get the current session from memcache
	item, session_id, err := getSession(request)
	json.Unmarshal(item.Value, &user)
	session.User = user
	session.Session_id = session_id

	//user information is updated
	if request.Method == "POST" {
		firstname := request.FormValue("firstname")
		lastname  := request.FormValue("lastname")
		email     := request.FormValue("email")

		password1 := request.FormValue("password1")
		password2 := request.FormValue("password2")

		//err is from getSession and means that getSession failed to retrieve
		//user info from memcache, if so then there is no way to retrieve the user info
		//anywhere. Why? because we lost the session id in cookie or from the url and 
		//there is no way to generate a key that can be use to find the user in datastore.
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In profile, user is impossible to find: %v ***", err)
			logout(response, request)

		//user info is in memcache
		}else{
			json.Unmarshal(item.Value, &user)
		}

		//if user decides to change email then check the datastore if the 
		//new email is unique.
		if user.Email != email {
			//perform a query to the datastore with the given new email
			//if the query returns 1 or more query results with the given
			//email, then it means new email is not unique.
			q := datastore.NewQuery("Users").Filter("Email =", email)
			i, _ := q.Count(ctx)

			//if the query count returns a non 0 result, it means that 
			//the email is already taken and therefore not unique
			if i > 0{
				log.Infof(ctx, "*** Error Info: In profile, email not unique ***")
				//if the user email is already in datastore then generate an error message 
				//and pass it to profile.html to show to the user.
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
			//and pass it to profile.html to show to the user.
			tpl.ExecuteTemplate(response, "profile.html", session)
			return
		}

		//below code will be reached if:
		//  - user has a session in memcache
		//	- new email is unique
		//  - user did not change email
		//  - password confirmations matches
		oldEmail := user.Email //save the old email so we can use it to find the current user in datastore

		//update the user fields
		user.Email = email    
		user.FirstName = firstname
		user.LastName = lastname

		//although password inputs are hidden in the profile.html it still goes to the server
		//via the POST method. if the user did not change password, we will still get a ""(empty string)
		//from the request.FormValue(...) and will be set as the new password of the user. so below if
		//statement guards for that instance and makes sure that the user will not have an empty password.
		if password1 != "" && password2 != ""{
			hashed_password, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
			if err != nil {
			//server error
				log.Errorf(ctx, "*** Error Debug: In profile, password hashing: %v ***", err)
				http.Error(response, err.Error(), 500)
				return
			}
			user.Password = string(hashed_password)
		}
		
		json, err := json.Marshal(user)
		if err != nil {
			//error marshalling user
			log.Errorf(ctx, "*** Error Debug: In profile json.Marshal: %v ***", err)
			//http.Error() replies to the request with the specified error message and HTTP code. 
			//The error message should be plain text.
			http.Error(response, err.Error(), 500)
			return
		}

		m := memcache.Item{
			Key:   session_id,
			Value: json,
		}
		memcache.Set(ctx, &m)

		//make a datastore query to find the current user(use the old email since the info of user in 
		//datastore is not yet updated.)
		//this query will return a slice of keys that has the user's oldEmail
		//since we already guaranteed that users will not have the same email at the same time
		//then this query will only return a slice of keys with one element in it.
		q := datastore.NewQuery("Users").Filter("Email =", oldEmail).KeysOnly()
		keys, _ := q.GetAll(ctx, nil)

		//keys[0] because it is guaranteed that the query will only return a slice of keys
		//with only 1 key in it.
		_, err = datastore.Put(ctx, keys[0], &user)
		if err != nil {
			//server error
			log.Errorf(ctx, "*** Error Debug: In profile, failed to save updated user to datastore: %v ***", err)
			http.Error(response, err.Error(), 500)
			return
		}

		//redirect so the page will refresh and the information of the user in profile.html will
		//be updated.
		http.Redirect(response, request, "/profile?id="+session.Session_id, http.StatusSeeOther)
	}

	tpl.ExecuteTemplate(response, "profile.html", session)
}


//files handler just serves the files.html
//the action happens in filehelper handler which handles ajax requests from file.js
func files(response http.ResponseWriter, request *http.Request) {
	var user User
	var session Session
	//get the session 
	ctx := appengine.NewContext(request)
	item, session_id, err := getSession(request)
	json.Unmarshal(item.Value, &user)
	session.User = user
	session.Session_id = session_id

	if err != nil{
		//no session found
		log.Errorf(ctx, "*** Error Debug: In files, user is impossible to find: %v ***", err)
		logout(response, request)
	}

	tpl.ExecuteTemplate(response, "files.html", session)	
}


//handles ajax request sent by files.js
func filehelper(response http.ResponseWriter, request *http.Request) {
	var user User
	var session Session
	//get the session
	ctx := appengine.NewContext(request)
	item, session_id, err := getSession(request)
	json.Unmarshal(item.Value, &user)
	session.User = user
	session.Session_id = session_id

	if err != nil{
		//no session found
		log.Errorf(ctx, "*** Error Debug: In files, user is impossible to find: %v ***", err)
		logout(response, request)
	}

	//create a new object handler
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "*** Error Debug: In files, storage.NewClient: %s", err)
		session.Message = "Oooops! Something went wrong try again"
		tpl.ExecuteTemplate(response, "files.html", session)
		return
	}
	defer client.Close()

	//query the gcs so we could get all the files of the LOGGED IN user
	//query delimiter is user.Id/
	query := &storage.Query{ Prefix:  strconv.Itoa(int(user.Id)) + "/" }
	objs, _ := client.Bucket(gcsBucket).List(ctx, query) //return a list of file objects

	//GET method renders all the files to the browser
	if request.Method == "GET" {
		var files_list []File
		for _, obj := range objs.Results {
			fileName := strings.TrimPrefix(obj.Name, strconv.Itoa(int(user.Id)) + "/")
			file := File{
				Name: fileName,
				//build the file link manually. Don't know if good practice, but heck it works.
				Source_Link: "https://storage.googleapis.com/" + gcsBucket + "/" + strconv.Itoa(int(user.Id)) + "/" + fileName,
				Download_Link: obj.MediaLink,
			}
			files_list = append(files_list, file)
		}
		
		//send the files_list to files.js
		err = json.NewEncoder(response).Encode(files_list)
	}

	//delete a file
	if request.Method == "DELETE" {
		filename := request.FormValue("filename") //filename is passed and retrieve here from the url
		//again build the filename in the form user.Id/filename and delete it
		err = client.Bucket(gcsBucket).Object( strconv.Itoa(int(user.Id)) + "/" + filename).Delete(ctx)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In filehelper, Delete: %s", err)
		}
		//any string here
		io.WriteString(response, "done")
	}
}
//go get github.com/gorilla/mux
//the session_id == uuid == cookie.Value is being passed in the url res]