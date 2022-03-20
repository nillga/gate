package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/controller"
	"github.com/go-chi/chi"
	"github.com/swaggo/http-swagger"
	_ "github.com/nillga/gate/docs"
)

var (
	jwtController controller.ApiGatewayController = controller.NewApiGatewayController()
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	cr := chi.NewRouter()

	cr.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
	))

	go log.Fatalln(http.ListenAndServe(os.Getenv("SWAG"), cr))

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