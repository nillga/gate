package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/nillga/api-gateway/controller"
	_ "github.com/nillga/api-gateway/docs"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	gatewayController controller.FrontendGatewayController = controller.NewApiGatewayController()
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
// @BasePath  /

func main() {
	cr := chi.NewRouter()

	cr.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
	))

	r := mux.NewRouter()
	
	// frontend takes bearer logic with the generated full value cookie
	// plan for API: bearer logic with reducable scope tokens --> security, somewhat
	r.HandleFunc("/user/signup", gatewayController.SignUp)
	r.HandleFunc("/user/login", gatewayController.Login)
	r.HandleFunc("/user/logout", gatewayController.Logout)
	r.HandleFunc("/user/delete", gatewayController.Delete)
	r.HandleFunc("/user", gatewayController.GetUser)
	r.HandleFunc("/mehms", gatewayController.Mehms)
	r.HandleFunc("/mehms/add", gatewayController.Add)
	r.HandleFunc("/mehms/{id}", gatewayController.SpecificMehm)
	r.HandleFunc("/mehms/{id}/like", gatewayController.LikeMehm)
	r.HandleFunc("/mehms/{id}/remove", gatewayController.Remove)
	r.HandleFunc("/mehms/{id}/update", gatewayController.EditMehm)
	r.HandleFunc("/comments/new", gatewayController.NewComment)
	r.HandleFunc("/comments/get/{id}", gatewayController.GetComment)
	r.HandleFunc("/comments/update", gatewayController.EditComment)
	r.HandleFunc("/comments/remove", gatewayController.DeleteComment)

	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Credentials", "Cookie"},
	})
	l := log.Logger{}
	l.SetOutput(os.Stdout)
	c.Log = &l

	go func() {
		log.Fatalln(http.ListenAndServe(os.Getenv("SWAG"), c.Handler(cr)))
	}()
	log.Fatalln(http.ListenAndServe(os.Getenv("PORT"), c.Handler(r)))
}