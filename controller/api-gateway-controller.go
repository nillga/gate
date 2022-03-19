package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/cache"
	"github.com/nillga/api-gateway/service"
	"github.com/nillga/jwt-server/entity"
	"io"
	"net/http"
	"os"
	"time"
)

type UserGateway interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
}

type MehmGateway interface {
	Mehms(w http.ResponseWriter, r *http.Request)
	Add(w http.ResponseWriter, r *http.Request)
	Remove(w http.ResponseWriter, r *http.Request)
	LikeMehm(w http.ResponseWriter, r *http.Request)
	SpecificMehm(w http.ResponseWriter, r *http.Request)
}

type CommentGateway interface {
	GetComment(w http.ResponseWriter, r *http.Request)
	NewComment(w http.ResponseWriter, r *http.Request)
}

type ApiGatewayController interface {
	UserGateway
	MehmGateway
	CommentGateway
}

type controller struct {
}

func NewApiGatewayController() ApiGatewayController {
	return &controller{}
}

var (
	gatewayService = service.NewService()
	gatewayCache   = cache.NewCache()
	userGateway    = os.Getenv("USERS_HOST")
	mehmGateway    = os.Getenv("MEHMS_HOST")
)

// gateway/user/signup
func (c *controller) SignUp(w http.ResponseWriter, r *http.Request) {
	proxiedReq, err := http.NewRequest(r.Method, userGateway+"/signup", r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(proxiedReq)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}

	cookie := res.Cookies()[0]
	var user *entity.User

	if err := json.NewDecoder(res.Body).Decode(user); err != nil {
		// manage
	}

	http.SetCookie(w, cookie)
	gatewayCache.Put(cookie.Value, user)
}

// gateway/user/login
func (c *controller) Login(w http.ResponseWriter, r *http.Request) {
	proxiedReq, err := http.NewRequest(r.Method, userGateway+"/login", r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(proxiedReq)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}

	// Login was successful
	var signedIn entity.User
	if err = json.NewDecoder(res.Body).Decode(&signedIn); err != nil {
		// manage
	}
	cooker, err := gatewayService.BuildCooker(&signedIn)
	if err != nil {
		// manage
	}

	gatewayCache.Put(cooker.Value, &signedIn)
	http.SetCookie(w, cooker)
}

// gateway/user/logout
func (c *controller) Logout(w http.ResponseWriter, r *http.Request) {
	cooker, err := r.Cookie("jwt")
	if err != nil {
		// manage
	}

	gatewayCache.Clear(cooker.Value)

	deadCookie := &http.Cookie{
		Name:    "jwt",
		Value:   "",
		Expires: time.Now().Add(time.Hour * (-2)),
		Path:    "/",
	}

	http.SetCookie(w, deadCookie)
}

// gateway/user/delete
func (c *controller) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := Auth(r)
	if err != nil {
		// manage
	}
	var deleteId *entity.DeleteUserInput
	if err = json.NewDecoder(r.Body).Decode(deleteId); err != nil {
		// manage
	}

	if !user.Admin && user.Id != deleteId.Id {
		// manage
	}

	pr, err := http.NewRequest(r.Method, userGateway+"/delete", r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}
}

// gateway/user
func (c *controller) GetUser(w http.ResponseWriter, r *http.Request) {
	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pr, err := http.NewRequest("GET", userGateway+"/resolve?id="+user.Id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// ----------------------

// gateway/mehms
func (c *controller) Mehms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	proxiedReq, err := http.NewRequest(r.Method, mehmGateway+"/mehms?"+r.URL.Query().Encode(), r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(proxiedReq)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// gateway/mehms/{id}
func (c *controller) SpecificMehm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		// error
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/"+id, r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// gateway/mehms/{id}
func (c *controller) LikeMehm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		// error
	}
	user, err := Auth(r)
	if err != nil {
		//handle
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/"+id+"/like?userId="+user.Id, r.Body) // tbd
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		// manage
	}
	if res.StatusCode != http.StatusOK {
		// manage
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// gateway/mehms/add
func (c *controller) Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/add", r.Body)
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

// gateway/mehms/remove
func (c *controller) Remove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Admin {
		pr, err := http.NewRequest("POST", mehmGateway+"/mehms/remove", r.Body)
		res, err := (&http.Client{}).Do(pr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if _, err = io.Copy(w, res.Body); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		return
	}

	var id *map[string]string

	if err = json.NewDecoder(r.Body).Decode(id); err != nil {
		// manage
	}

	mr, err := http.NewRequest("GET", mehmGateway+"/mehms/"+(*id)["id"], r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(mr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err = json.NewDecoder(res.Body).Decode(id); err != nil {
		// manage
	}
	if user.Username != (*id)["authorName"] {
		// manage
	}

	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/remove", r.Body)
	res, err = (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	return
}

// ---------------------

// gateway/comments/{id}
func (c *controller) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		// error
	}

	pr, err := http.NewRequest("GET", mehmGateway+"/comments/"+id, r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		// manage
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// gateway/comments/new
func (c *controller) NewComment(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("comment") || !params.Has("mehmId") {
		// manage
	}

	user, err := Auth(r)
	if err != nil {
		// manage
	}

	pr, err := http.NewRequest("GET", mehmGateway+"/comments/new?"+r.URL.Query().Encode()+"&userId="+user.Id, r.Body)
	if err != nil {
		// manage
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		// manage
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		// manage
	}
}

// ---------------------

func Auth(r *http.Request) (*entity.User, error) {
	// do auth
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, err
	}

	if user, inCache := gatewayCache.Get(cookie.Value); inCache {
		// modify request
		// ==> add req param id
		return user, nil
	}

	user, err := gatewayService.ReadCooker(cookie)
	if err != nil {
		return nil, err
	}
	gatewayCache.Put(cookie.Value, user)
	return user, nil
}