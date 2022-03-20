package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/cache"
	"github.com/nillga/api-gateway/service"
	"github.com/nillga/jwt-server/entity"
	"github.com/nillga/jwt-server/errors"
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

// SignUp godoc
// @Summary      Used to register a new user
// @Description  Requires the user's credentials: namely their nickname, email and password
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input   body      entity.SignupInput  true  "Input data"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user/signup [post]
func (c *controller) SignUp(w http.ResponseWriter, r *http.Request) {
	_ = errors.ProceduralError{}
	pr, err := http.NewRequest(r.Method, userGateway+"/signup", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	var user entity.User

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	cookie, err := gatewayService.BuildCooker(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cookie err")
		return
	}

	http.SetCookie(w, cookie)
	gatewayCache.Put(cookie.Value, &user)
}

// Login godoc
// @Summary      Used to login and receive a JWT
// @Description  Identifier id can be email or username
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input   body      entity.LoginInput  true  "Input data"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user/login [get]
func (c *controller) Login(w http.ResponseWriter, r *http.Request) {
	pr, err := http.NewRequest(r.Method, userGateway+"/login", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("pr: ", err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("prDone", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	var user entity.User

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("json decode ", err)
		return
	}

	cookie, err := gatewayService.BuildCooker(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("cookie err")
		return
	}
	http.SetCookie(w, cookie)
	gatewayCache.Put(cookie.Value, &user)
}

// Logout godoc
// @Summary      Used to logout and remove a JWT
// @Description  Identifier id can be email or username
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user/logout [get]
func (c *controller) Logout(w http.ResponseWriter, r *http.Request) {
	cooker, err := r.Cookie("jwt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
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

// Delete godoc
// @Summary      Deletes a targeted User
// @Description  Self-delete; admins can delete anybody
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input   body      entity.DeleteUserInput  true  "Input data"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user/delete [delete]
func (c *controller) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var deleteId entity.DeleteUserInput
	if err = json.NewDecoder(r.Body).Decode(&deleteId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if !user.Admin && user.Id != deleteId.Id {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("rights", err)
		return
	}

	pr, err := http.NewRequest(r.Method, userGateway+"/delete?id="+deleteId.Id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	if user.Id != deleteId.Id {
		return
	}

	cooker, err := r.Cookie("jwt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
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

// GetUser godoc
// @Summary      Receive Info about ones self
// @Description  Password isnt cleared yet UwU
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  entity.User{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /user [get]
func (c *controller) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pr, err := http.NewRequest("GET", userGateway+"/resolve?id="+user.Id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// ----------------------

// GetMehms godoc
// @Summary      Returns a page of mehms
// @Description  Pagination can be handled via query params
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        skip   query      int  false  "How many mehms will be skipped"
// @Param        take   query      int  false  "How many mehms will be taken"
// @Success      200  {object}  map[string]dto.MehmDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms [get]
func (c *controller) Mehms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pr, err := http.NewRequest(r.Method, mehmGateway+"/mehms?"+r.URL.Query().Encode(), r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	for k, v := range res.Header {
		log.Println(k, ":", v)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Println(bodyString)

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// GetSpecificMehm godoc
// @Summary      Returns a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  dto.MehmDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms/{id} [get]
func (c *controller) SpecificMehm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := Auth(r)
	if err == nil {
		id += "?userId=" + user.Id
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/mehms/get/"+id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// LikeMehm godoc
// @Summary      Used to like a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms/{id}/like [post]
func (c *controller) LikeMehm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/mehms/"+id+"/like?userId="+user.Id, r.Body) // tbd
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println(pr.URL.RawPath)
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// AddMehm godoc
// @Summary      Uploads a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        id   formData      int  true  "The ID of the requested mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms/add [post]
func (c *controller) Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/add", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
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

// RemoveMehm godoc
// @Summary      Used to delete a specified mehm
// @Description  optionally showing info for privileged user
// @Tags         mehms
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /mehms/{id}/remove [post]
func (c *controller) Remove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	adminString := "false"

	if user.Admin {
		adminString = "true"
	}

	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/"+id+"/remove?userId="+user.Id+"&admin="+adminString, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		return
	}
	pr.Header.Set("Content-Type", "application/json")
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		w.WriteHeader(res.StatusCode)
		log.Println(err)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		return
	}
}

// ---------------------

// GetComment godoc
// @Summary      Used to show a specified comment
// @Description  optionally showing info for privileged user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "The ID of the requested mehm"
// @Success      200  {object}  dto.CommentDTO{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /comments/get/{id} [get]
func (c *controller) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pr, err := http.NewRequest("GET", mehmGateway+"/comments/get/"+id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// AddComment godoc
// @Summary      Used to add a new comment
// @Description  optionally showing info for privileged user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment   query      string  true  "The comment"
// @Param        mehmId   query      int  true  "The mehm"
// @Success      200  {object}  interface{}
// @Failure      400  {object}  errors.ProceduralError
// @Failure      404  {object}  errors.ProceduralError
// @Failure      500  {object}  errors.ProceduralError
// @Router       /comments/get/{id} [get]
func (c *controller) NewComment(w http.ResponseWriter, r *http.Request) {
	log.Println("new comment!!")
	params := r.URL.Query()
	if !params.Has("comment") || !params.Has("mehmId") {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("oof")
		return
	}

	user, err := Auth(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	pr, err := http.NewRequest(r.Method, mehmGateway+"/comments/new?"+r.URL.Query().Encode()+"&userId="+user.Id, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
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
