package GodoersToDo

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"encoding/json"
	"google.golang.org/appengine/log"
	"time"
)


//gets the session in memcache
func getSession(request *http.Request) (*memcache.Item, string, error) {
	//NewContext returns a context for an in-flight HTTP request. This function is cheap.
	ctx := appengine.NewContext(request)

	//get cookie
	cookie, err := request.Cookie("session-info")

	//if cookie does not exists, return try the url
	if err != nil {
		//check for session id in url ../?id="session_id"
		key := "id"
		val := request.URL.Query().Get(key)
		//if there is an id in url then check the memcache using that id
		if val != "" {
			item, err := memcache.Get(ctx, val)
			//not found in memcache
			if err != nil {
				return &memcache.Item{}, "", err 
			}
			return item, val, nil
		}
		return &memcache.Item{}, "", err
	}

	//cookie exists
	//retrieve from memcache
	item, err := memcache.Get(ctx, cookie.Value)
	//if cookie not in memcache
	if err != nil {
		return &memcache.Item{}, "", err
	}
	//else
	return item, cookie.Value, nil
}


func createSession(response http.ResponseWriter, request *http.Request, user User) string{
	ctx := appengine.NewContext(request)

	id, _ := uuid.NewV4() //generate new uuid

	//create a cookie with uuid
	cookie := &http.Cookie{
		Name:   "session-info",
		Value:  id.String(),
		//MaxAge is the amount of time before a cookie is deleted(seconds).
		//In this case cookie will expire in an hour once a session is started.
		MaxAge: 60 * 60,
		//Uncomment next 2 lines before deploying
		// Secure: true,
		// HttpOnly: true,
	}

	http.SetCookie(response, cookie)

	//After creating cookie session-info, let's save it to memcache together with user info
	json, err := json.Marshal(user)
	if err != nil {
		//error marshalling user
		log.Errorf(ctx, "*** Error Debug: In createSession json.Marshal: %v ***", err)
		//http.Error() replies to the request with the specified error message and HTTP code. 
		//The error message should be plain text.
		http.Error(response, err.Error(), 500)
		return ""
	}
	//for debugging purposes: paste the cookie id from the terminal to memcache viewer
	//to see if the user(json) is being cached in memcache
	//log.Infof(ctx, "Cookie Id:" + " " + cookie.Value)
	log.Infof(ctx, cookie.Value)
	m := memcache.Item{
		Key:   cookie.Value,
		Value: json,
	}
	memcache.Set(ctx, &m)

	

	return cookie.Value
}



func deleteSession(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	var session Session
	_, session_id, err := getSession(request)
	session.Session_id = session_id

	cookie, err := request.Cookie("session-info")
	//no cookie

	if err != nil {
		item := memcache.Item{
			Key:   session.Session_id,
			Value: []byte(""),
			Expiration: time.Duration(1 * time.Microsecond),
		}
		memcache.Set(ctx, &item)
		return
	}
	cookie.MaxAge = -1
	http.SetCookie(response, cookie)
	item := memcache.Item{
		Key:   cookie.Value,
		Value: []byte(""),
		Expiration: time.Duration(1 * time.Microsecond),
	}
	memcache.Set(ctx, &item)
}

