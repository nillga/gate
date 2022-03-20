package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/controller"
)

var (
	jwtController controller.ApiGatewayController = controller.NewApiGatewayController()
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/user/signup", jwtController.SignUp)
	r.HandleFunc("/user/login", jwtController.Login)
	r.HandleFunc("/user/logout", jwtController.Logout)
	r.HandleFunc("/user/delete", jwtController.Delete)
	r.HandleFunc("/user", jwtController.GetUser)
	r.HandleFunc("/mehms", jwtController.Mehms)
	r.HandleFunc("/mehms/add", jwtController.Add)
	r.HandleFunc("/mehms/{id}", jwtController.SpecificMehm)
	r.HandleFunc("/mehms/{id}/like", jwtController.LikeMehm)
	r.HandleFunc("/mehms/{id}/remove", jwtController.Remove)
	r.HandleFunc("/comments/new", jwtController.NewComment)
	r.HandleFunc("/comments/get/{id}", jwtController.GetComment)

	log.Fatalln(http.ListenAndServe(os.Getenv("PORT"), r))
}