package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/dto"
	"github.com/nillga/api-gateway/service"
	"github.com/nillga/api-gateway/utils"
	"github.com/nillga/jwt-server/entity"
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

type FrontendGatewayController interface {
	UserGateway
	MehmGateway
	CommentGateway
}

type controller struct {
}

func NewApiGatewayController() FrontendGatewayController {
	return &controller{}
}

var (
	gatewayService = service.NewService()
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
	w.Header().Set("Content-Type", "application/json")
	pr, err := http.NewRequest(r.Method, userGateway+"/signup", r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
	}
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
// @Router       /user/login [post]
func (c *controller) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pr, err := http.NewRequest(r.Method, userGateway+"/login", r.Body)
	pr.Header.Set("Content-Type", "application/json")
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	var user entity.User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	cookie, err := gatewayService.BuildCooker(&user)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	if err := json.NewEncoder(w).Encode(cookie); err != nil {
		utils.InternalServerError(w, err)
		return
	}
	gatewayService.Cache(cookie.Value, &user)
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
	w.Header().Set("Content-Type", "application/json")
	jwt, err := gatewayService.ReadBearer(r.Header.Get("Authorization"))
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	gatewayService.UnCache(jwt)

	utils.DeleteJwtCookie(w)
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
	w.Header().Set("Content-Type", "application/json")
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}
	var deleteId entity.DeleteUserInput
	if err = json.NewDecoder(r.Body).Decode(&deleteId); err != nil {
		utils.UnprocessableEntity(w, err)
		return
	}

	if !user.Admin && user.Id != deleteId.Id {
		utils.Forbidden(w, err)
		return
	}

	pr, err := http.NewRequest(r.Method, userGateway+"/delete?id="+deleteId.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	jwt, err := gatewayService.ReadBearer(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	gatewayService.UnCache(jwt)

	if user.Id != deleteId.Id {
		return
	}

	utils.DeleteJwtCookie(w)
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
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}
	pr, err := http.NewRequest("GET", userGateway+"/resolve?id="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	user, err := gatewayService.Auth(r)
	if err == nil {
		id += "?userId=" + user.Id
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/mehms/get/"+id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}
	pr, err := http.NewRequest(r.Method, mehmGateway+"/mehms/"+id+"/like?userId="+user.Id, r.Body) // tbd
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}
	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/add?userId="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	pr.Header["Content-Type"] = r.Header["Content-Type"]

	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}
	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.BadRequest(w, err)
		return
	}

	adminString := "false"

	if user.Admin {
		adminString = "true"
	}

	pr, err := http.NewRequest("POST", mehmGateway+"/mehms/"+id+"/remove?userId="+user.Id+"&admin="+adminString, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	pr.Header.Set("Content-Type", "application/json")
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.BadRequest(w, fmt.Errorf("comment specification went wrong"))
		return
	}

	pr, err := http.NewRequest("GET", mehmGateway+"/comments/get/"+id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
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
		utils.BadRequest(w, fmt.Errorf("request query parameters must be comment AND mehmId"))
		return
	}

	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	pr, err := http.NewRequest(r.Method, mehmGateway+"/comments/new?"+r.URL.Query().Encode()+"&userId="+user.Id, r.Body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

func (c *controller) EditComment(w http.ResponseWriter, r *http.Request) {
	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	if !user.Admin {
		utils.Forbidden(w, fmt.Errorf("not authorized"))
		return
	}

	var input dto.CommentInput

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.UnprocessableEntity(w, fmt.Errorf("format problems"))
		return
	}

	if input.UserId, err = strconv.ParseInt(user.Id, 10, 64); err != nil {
		utils.InternalServerError(w, fmt.Errorf("could not resolve user"))
		return
	}
	input.Admin = user.Admin

	body := bytes.NewBuffer([]byte{})

	if err = json.NewEncoder(body).Encode(input); err != nil {
		utils.InternalServerError(w, fmt.Errorf("failed repeating request"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehmGateway+"/comments/update", body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}

func (c *controller) EditMehm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		utils.BadRequest(w, fmt.Errorf("mehm specification went wrong"))
		return
	}

	user, err := gatewayService.Auth(r)
	if err != nil {
		utils.Unauthorized(w, err)
		return
	}

	if !user.Admin {
		utils.Forbidden(w, fmt.Errorf("not authorized"))
		return
	}

	var input dto.MehmInput

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.UnprocessableEntity(w, fmt.Errorf("format problems"))
		return
	}

	if input.UserId, err = strconv.ParseInt(user.Id, 10, 64); err != nil {
		utils.InternalServerError(w, fmt.Errorf("could not resolve user"))
		return
	}
	input.Admin = user.Admin

	body := bytes.NewBuffer([]byte{})

	if err = json.NewEncoder(body).Encode(input); err != nil {
		utils.InternalServerError(w, fmt.Errorf("failed repeating request"))
		return
	}

	pr, err := http.NewRequest(r.Method, mehmGateway+"/"+id+"/update", body)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}
	res, err := (&http.Client{}).Do(pr)
	if err != nil {
		utils.BadGateway(w, err)
		return
	}
	if res.StatusCode != http.StatusOK {
		utils.WrongStatus(w, res)
		return
	}

	if _, err = io.Copy(w, res.Body); err != nil {
		utils.InternalServerError(w, err)
	}
}
