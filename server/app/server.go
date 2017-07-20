package app

import (
	"net/http"

	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/expense"
	"github.com/smalex-als/health-tracker/server/meal"
	"github.com/smalex-als/health-tracker/server/user"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	gin.SetMode(gin.ReleaseMode)

	// Starts a new Gin instance with no middle-ware
	r := gin.New()
	// Global middleware
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "static")

	register := &common.HandlerRegister{r}
	services := []common.ServiceRegister{
		&user.AuthRemoteService{},
		&user.UserRemoteService{},
		&expense.ExpenseRemoteService{},
		&meal.MealRemoteService{},
	}
	for _, v := range services {
		v.Register(register)
	}

	r.GET("/install", handleInstall)
	r.GET("/", handleIndex)
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Handle all requests using net/http
	http.Handle("/", r)
}

func handleInstall(c *gin.Context) {
	ctx := appengine.WithContext(c, c.Request)
	host := c.Request.Host
	client := urlfetch.Client(ctx)
	for _, module := range []string{"users", "expenses", "meals"} {
		resp, err := client.Get("http://" + host + "/v1/" + module + "-install/")
		if err != nil {
			panic(err)
		}
		log.Infof(ctx, "install %s status %v", module, resp.Status)
	}
	c.String(200, "OK")
}

func handleIndex(c *gin.Context) {
	common.InitContext(c)
	ctx := common.GetAppEngineContext(c)
	user := common.UserFromContext(ctx)
	h := gin.H{}
	if user != nil {
		h["User"] = user
	}
	c.HTML(200, "index.html", h)
	// w := c.Writer
	// p, ok := w.(http.Pusher)
	// if ok {
	// 	log.Infof(ctx, "http2 push")
	// 	p.Push("/static/js/admin.nocache.js", nil)
	// 	p.Push("/static/css/bootstrap.min.css", nil)
	// 	p.Push("/static/css/ie10-viewport-bug-workaround.css", nil)
	// 	p.Push("/static/css/signin.css", nil)
	// }
}
