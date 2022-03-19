package main

import (
	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/controller"
	"github.com/nillga/jwt-server/repository"
	"github.com/nillga/jwt-server/service"
	"net/http"
	"os"
)

var (
	jwtRepo       repository.JwtRepository        = repository.NewPostgresRepo()
	jwtService    service.JwtService              = service.NewJwtService(jwtRepo)
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
	r.HandleFunc("/mehms/remove", jwtController.Remove)
	r.HandleFunc("/mehms/{id}", jwtController.SpecificMehm)
	r.HandleFunc("/mehms/{id}/like", jwtController.LikeMehm)
	r.HandleFunc("/comments/{id}", jwtController.GetComment)
	r.HandleFunc("/comments/new", jwtController.NewComment)

	http.ListenAndServe(os.Getenv("PORT"), r)
}
