package GodoersToDo

import (
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	//"google.golang.org/appengine/memcache"
	"encoding/json"
	"google.golang.org/cloud/storage"
	"io"
	"strings"
	"strconv"
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
	
	if request.Method == "GET" {
		q := datastore.NewQuery("Todos").Filter("UserId =", user.Id)
		iterator := q.Run(ctx)
		todos := make([]ToDo, 0)
		for{
			var todo ToDo
			
			key, err := iterator.Next(&todo)
			if err == datastore.Done {
				break
			} else if err != nil {
				log.Errorf(ctx, "*** Error Debug: In dashboard, retrieving todos: %v ***", err)
				http.Error(response, err.Error(), 500)
				return
			}
			todo.ToDoId = key.IntID()
			todos = append(todos, todo)
		}

		err = json.NewEncoder(response).Encode(todos)
		if err != nil {
			log.Errorf(ctx, "*** Error Debug: In dashboard, jsonifying? todos: %v ***", err)
			return
		}
	}

	if request.Method == "POST" {
		request.ParseMultipartForm(10000000000)
		content := request.FormValue("content")
		src, hdr, err := request.FormFile("file")
		if err == nil {		
			defer src.Close()

			//only allow jpeg, jpg or png files
			ext := hdr.Filename[strings.LastIndex(hdr.Filename, ".")+1:]
			switch ext {
				case "jpg", "jpeg", "png":
					
				default:
					log.Infof(ctx, "*** Error Info: In dashboard, we only accept .jpeg, .jpg or .png files ***")
					session.Message = "Only files with extensions .jpeg, .jpg or .png files are accepted"
					tpl.ExecuteTemplate(response, "dash.html", session)
					return
			}

			fileName = strconv.Itoa(int(user.Id)) + "/" + hdr.Filename
			
			client, err := storage.NewClient(ctx)
			if err != nil {
				log.Errorf(ctx, "*** Error Debug: In dashboard, storage.NewClient: %s", err)
				session.Message = "Oooops! Something went wrong try again"
				tpl.ExecuteTemplate(response, "dash.html", session)
				return
			}
			defer client.Close()

			writer := client.Bucket(gcsBucket).Object(fileName).NewWriter(ctx)
			writer.ACL = []storage.ACLRule{
				{storage.AllUsers, storage.RoleReader},
			}
			
			io.Copy(writer, src)
			err = writer.Close()
			if err != nil {
				log.Errorf(ctx, "*** Error Debug: In dashboard, writer.Close: %s", err)
				session.Message = "Oooops! Something went wrong try again"
				tpl.ExecuteTemplate(response, "dash.html", session)
				return
			}
		}

		todo := ToDo{
			UserId:  user.Id, 
			Content: content,
			Photo:   fileName,
		}
		key := datastore.NewIncompleteKey(ctx, "Todos", nil)
		key, err = datastore.Put(ctx, key, &todo)
	}

}