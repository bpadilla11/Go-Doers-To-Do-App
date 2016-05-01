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


//gets the session in memcache using the cookie value or value from url as the key
func getSession(request *http.Request) (*memcache.Item, string, error) {
	//NewContext returns a context for an in-flight HTTP request. This function is cheap.
	ctx := appengine.NewContext(request)

	//get cookie
	cookie, err := request.Cookie("session-info")

	//if cookie does not exists, try the url
	if err != nil {
		//check for session id in url ../?id="session_id"
		key := "id"
		val := request.URL.Query().Get(key) //get value from url

		//if there is an id in url then check the memcache using that id
		if val != "" {
			item, err := memcache.Get(ctx, val)
			//not found in memcache, no session
			if err != nil {
				return &memcache.Item{}, "", err 
			}
			return item, val, nil //session found in memcache return the memcache item, the value in url, nil
		}
		return &memcache.Item{}, "", err //error
	}

	//cookie exists
	//retrieve from memcache with key as cookie.Value
	item, err := memcache.Get(ctx, cookie.Value)
	//if cookie not in memcache
	if err != nil {
		return &memcache.Item{}, "", err //no  user session in memcache
	}
	//session found in memcache return the memcache item, the cookie.Value, nil
	return item, cookie.Value, nil
	//this function also returns the cookie.Value or the value got from the url so we can add it
	//to the url to maintain session even without a cookie

	//I don't know if this is good practice, but to show that a session can be maintained even by just
	//the value from url
}


//creates a new session for users that logs in or registers
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
		//Secure: true,
		//HttpOnly: true,
	}

	http.SetCookie(response, cookie)

	//After creating cookie session-info, let's save it to memcache together with user info
	//convert user info to json that will be used as the mecache item.
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
	//log.Infof(ctx, cookie.Value)

	//save the user info to memcache with the cookie.Value(uuid) as key and the value
	//as the jsonified user info
	m := memcache.Item{
		Key:   cookie.Value,
		Value: json,
		Expiration: time.Duration(60 * time.Minute), //item in memcache will expire after an hour same as cookie
	}
	memcache.Set(ctx, &m)

	//returns the cookie.Value so it can be used as a value in url to maintain session even without
	//cookies
	return cookie.Value
}


//deletes the session by deleting the cookie in the browser and deleting the session
//in memcache either by getting the cookie.Value or the value from url
func deleteSession(response http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)
	var session Session
	//remember: getSession also returns the session_id from cookie.Value or value from url
	_, session_id, err := getSession(request)
	session.Session_id = session_id

	//get cookie
	cookie, err := request.Cookie("session-info")

	//no cookie
	if err != nil {
		//if their is no cookie in the browser then use the session_id from url
		//to reference the memcache item
		/*item := memcache.Item{
			Key:   session.Session_id,
			Value: []byte(""),
			Expiration: time.Duration(1 * time.Microsecond),
		}
		memcache.Set(ctx, &item)*/
		memcache.Delete(ctx, session_id)
		return
	}
	//setting this to -1 means delete now
	cookie.MaxAge = -1
	http.SetCookie(response, cookie)
	memcache.Delete(ctx, cookie.Value)
	//if cookie exists then use the cookie.Value to reference the memcache item
	/*item := memcache.Item{
		Key:   cookie.Value,
		Value: []byte(""),
		Expiration: time.Duration(1 * time.Microsecond),
	}
	memcache.Set(ctx, &item)*/
}

