package main

// @title           Blog REST API
// @version         1.0
// @description     Blog API with JWT auth, posts, and comments.
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	execute()
}
