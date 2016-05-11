package GodoersToDo

import (
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"encoding/json"
	"google.golang.org/cloud/storage"
	"io"
	"strings"
	"strconv"
	"time"
)


func todo(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
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
	var fileName string
	
	//method GET displays all the user's todo objects.
	if request.Method == "GET" {
		//query the datastore to get all the todo's of the LOGGED IN user and sort it according to data(oldest first)
		q := datastore.NewQuery("Todos").Filter("UserId =", user.Id).Order("Date")
		iterator := q.Run(ctx)

		todos := make([]ToDo, 0)

		//unlimited for loop because we have iterator, so we will know if we reach the last todo by checking if
		//iterator is done.
		for{
			var todo ToDo
			
			key, err := iterator.Next(&todo)
			if err == datastore.Done {
				break
			} else if err != nil {
				log.Errorf(ctx, "*** Error Debug: In todos, retrieving todos: %v ***", err)
				http.Error(response, err.Error(), 500)
				return
			}
			todo.ToDoId = key.IntID()
			//add the todo to the todos list so we can json it and pass it via ajax.
			todos = append(todos, todo)
		}

		//convert todos list to json and pass it to todo.js so it can be rendered to the browser
		err = json.NewEncoder(response).Encode(todos)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In todos, jsonifying? todos: %v ***", err)
			return
		}
	}

	if request.Method == "POST" {
		//ParseMultipartForm is used here because todo.js passes a form data which contains the todo content
		//and todo photo(sending a multipart form)
		request.ParseMultipartForm(10000000000)
		content := request.FormValue("content")
		src, hdr, err := request.FormFile("file")
		
		//there is a file uploaded 
		if err == nil {		
			defer src.Close()

			//only allow jpeg, jpg or png files
			ext := hdr.Filename[strings.LastIndex(hdr.Filename, ".")+1:]
			log.Infof(ctx, ext)
			if ext != "png" && ext != "jpg" && ext != "jpeg" {
				log.Infof(ctx, "*** Error Info: In todo, we only accept .jpeg, .jpg or .png files ***")
				session.Message = "Only files with extensions .jpeg, .jpg or .png files are accepted"
					
				//if the uploaded file is not jpg, jpeg or png then notify the todo.js
				//by checking the Photo_Link/Photo_Media string if it is invalid. Not the
				//best way but I am in a hurry and most importantly it works.
				todo := ToDo{
					UserId:  0, 
					Content: "",
					Status:  queued,
					Photo_Link: "invalid",
					Photo_Media:"invalid", 
				}
				err = json.NewEncoder(response).Encode(todo)
				return
			}

			//file is jpeg, jpg, or png
			//generate a filename in the form user.Id/filename so every file uploaded by the user
			//will be stored in a folder with the name as the user's id in gcs bucket.
			fileName = strconv.Itoa(int(user.Id)) + "/" + hdr.Filename
			
			client, err := storage.NewClient(ctx)
			if err != nil {
				log.Errorf(ctx, "*** Error Debug: In todo, storage.NewClient: %s", err)
				session.Message = "Oooops! Something went wrong try again"
				//tpl.ExecuteTemplate(response, "dash.html", session)
				return
			}
			defer client.Close()

			//get a object handle so we could save the file to gcs
			writer := client.Bucket(gcsBucket).Object(fileName).NewWriter(ctx)
			//set the acl rule so the file will have a public link.
			writer.ACL = []storage.ACLRule{
				{storage.AllUsers, storage.RoleReader},
			}
			io.Copy(writer, src) //copy the contents of the file to the object handle effectively saving it to gcs
			err = writer.Close()
			if err != nil {
				log.Errorf(ctx, "*** Error Debug: In todo, writer.Close: %s", err)
				session.Message = "Oooops! Something went wrong try again"
				//tpl.ExecuteTemplate(response, "dash.html", session)
				return
			}

			//after saving the object to gcs query the gcs so we could get the file we just uploaded to gcs
			//and get its mediaLink and set the Photo_Link and send that todo object to todo.js
			query := &storage.Query{ Prefix: fileName }
			objs, _ := client.Bucket(gcsBucket).List(ctx, query)

			var s string
			//in our query we explicitly put specific filename so we could get that specific file in gcs.
			for _, obj := range objs.Results {
				s = obj.MediaLink
			}

			//create a new todo object so we could save it to datastore with the file we got from gcs query
			todo := ToDo{
				UserId:  user.Id, 
				Content: content,
				Status:  queued,
				Date: time.Now().Format("Mon Jan 2 2006 03:04 PM"),
				Photo_Link: "https://storage.googleapis.com/" + gcsBucket + "/" + fileName,
				Photo_Media: s,
			}
			key := datastore.NewIncompleteKey(ctx, "Todos", nil)
			key, err = datastore.Put(ctx, key, &todo)
			todo.ToDoId = key.IntID()
			key, err = datastore.Put(ctx, key, &todo)
			err = json.NewEncoder(response).Encode(todo) //send it to todo.js
			return
		}

		//if user did not upload a file then leave the Photo_... blank.
		todo := ToDo{
				UserId:  user.Id, 
				Content: content,
				Status:  queued,
				Date: time.Now().Format("Mon Jan 2 2006 03:04 PM"),
				Photo_Link:   "",
				Photo_Media:  "",
		}

		//save it to datstore
		key := datastore.NewIncompleteKey(ctx, "Todos", nil)
		key, err = datastore.Put(ctx, key, &todo)
		todo.ToDoId = key.IntID()
		key, err = datastore.Put(ctx, key, &todo)
		err = json.NewEncoder(response).Encode(todo) //send it to todo.js
	}

	//user wants to delete a todo object
	if request.Method == "DELETE" {
		//todo.js delete method passes the todo's object id so we could delete it from
		//datastore. The id of todo is being passed in the url instead of normal input.
		todo_id, _ := strconv.ParseInt(request.FormValue("todo"), 10, 64)

		//get the todo object from datastore
		key := datastore.NewKey(ctx, "Todos", "", todo_id, nil)
		err = datastore.Delete(ctx, key)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In todo, Delete: %s", err)
		}
		//pass any string here
		io.WriteString(response, "done")
	}

	if request.Method == "UPDATE" {
		var todo ToDo
		todo_id, _ := strconv.ParseInt(request.FormValue("todo"), 10, 64)

		//get the todo object from datastore
		key := datastore.NewKey(ctx, "Todos", "", todo_id, nil)
		datastore.Get(ctx, key, &todo)
		if todo.Status == queued {
			todo.Status = done
		} else{
			todo.Status = queued
		}
		datastore.Put(ctx, key, &todo)
		//pass any string here
		io.WriteString(response, "done")
	}
}