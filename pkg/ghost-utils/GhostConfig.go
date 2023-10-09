package ghostutils

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
	"gopkg.in/yaml.v3"
)

type GhostConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Port        int    `yaml:"port"`
	SurrealDB   struct {
		URL        string `yaml:"surrealdb-url"`
		Username   string `yaml:"surrealdb-username"`
		Password   string `yaml:"surrealdb-password"`
		Database   string `yaml:"surrealdb-database"`
		Namespace  string `yaml:"surrealdb-namespace"`
	} `yaml:"surrealdb"`
	TailwindCSS struct {
		Input  string `yaml:"input"`
		Output string `yaml:"output"`
	} `yaml:"tailwindcss"`
}

// New returns a new GhostConfig struct 
// used to load the ghost.yaml file into a ghost project
//
// Example:
//  ghostConfig, err := ghostutils.New()
//  if err != nil { 
//      log.Fatal(err) 
//  }
//  fmt.Println(ghostConfig.Name)
//
// Returns:
//  GhostConfig struct
//  error
func New() (GhostConfig, error) {
    // load ghost config from the root of the project
	ghostConfig := GhostConfig{}
    ghostConfigFile, err := ioutil.ReadFile("./ghost.yaml")
	if err != nil {
    
		return ghostConfig, err
	}
	err = yaml.Unmarshal(ghostConfigFile, &ghostConfig)
	if err != nil {
		return ghostConfig, err
	}
	return ghostConfig, nil
}

// NewFromPath returns a new GhostConfig struct
// used to load the ghost.yaml file into a ghost project
// from a specified path instead of the root of the project.
// This is particularly useful when you want to load a ghost.testing.yaml
// file for testing purposes, or something similar.
//
// Example: 
//  ghostConfig, err := ghostutils.NewFromPath("./ghost.testing.yaml")
//  if err != nil {
//      log.Fatal(err)
//  }
//  fmt.Println(ghostConfig.Name)
//
// Returns:
//  GhostConfig struct
//  error
func NewFromPath(path string) (GhostConfig, error) {
    // load ghost config from the root of the project
    ghostConfig := GhostConfig{}
    ghostConfigFile, err := ioutil.ReadFile(path)
    if err != nil {
        return ghostConfig, err
    }
    err = yaml.Unmarshal(ghostConfigFile, &ghostConfig)
    if err != nil {
        return ghostConfig, err
    }
    return ghostConfig, nil
}

type GhostRoute interface {
    New(path string, db *surrealdb.DB) GhostRoute
    Route(rg *gin.RouterGroup)
    DB() *surrealdb.DB
    RG() *gin.RouterGroup
}

type BasicRoute struct {
    db *surrealdb.DB
    RouteGroup *gin.RouterGroup
    Path string
}

// New returns a new BasicRoute struct
// used to create a new route for a ghost project 
// using the surrealdb database. 
// 
// Arguments:
//  path: string
//  db: *surrealdb.DB 
//
// Example: 
//  basicRoute := BasicRoute{} 
//  db, err := ghostConfig.Setup(r)
//  if err != nil {
//      log.Fatal(err)
//  }
//  basicRoute.New("/basic", db)
//
// Returns:
//  BasicRoute struct
func (basicRoute BasicRoute) New(path string, db *surrealdb.DB) GhostRoute {
    return BasicRoute{
        db: db,
        Path: path,
        RouteGroup: nil,
    }
}

// DB returns the surrealdb database 
// used to create the route. 
// 
// Example: 
//  basicRoute := BasicRoute{} 
//  db, err := ghostConfig.Setup(r) 
//  if err != nil { 
//      log.Fatal(err) 
//  } 
//  basicRoute.New("/basic", db) 
//  _ = basicRoute.DB()) 
//
// Returns:
//  *surrealdb.DB
func (basicRoute BasicRoute) DB() *surrealdb.DB {
    return basicRoute.db
}

// Route is used to create a new route for a ghost project 
// using the surrealdb database. 
// 
// Arguments: 
//  rg: *gin.RouterGroup 
// 
// Example: 
//  api := BasicRoute{} 
//  db, err := ghostConfig.Setup(r) 
//  if err != nil { 
//      log.Fatal(err) 
//  } 
//  api.New("/api", db)
//  // setup routes for api using api.RG()
// 
func (basicRoute BasicRoute) Route(rg *gin.RouterGroup) {
    basic := rg.Group(basicRoute.Path)
    basicRoute.RouteGroup = basic
}

// RG returns the gin.RouterGroup used to create the route. 
// used in other parts of the ghost to create routes. 
// 
// Example:
//  basicRoute := BasicRoute{} 
//  db, err := ghostConfig.Setup(r) 
//  if err != nil { 
//      log.Fatal(err) 
//  } 
//  basicRoute.New("/basic", db)
//  basicRouteRoute := basicRoute.RG()
//  basicRouteRoute.GET("/", func(c *gin.Context) {
//      c.HTML(http.StatusOK, "index.html", gin.H{})
//  })
//  r.Run(fmt.Sprintf(":%d", ghostConfig.Port))
// 
// Returns:
//  *gin.RouterGroup
func (basicRoute BasicRoute) RG() *gin.RouterGroup {
    return basicRoute.RouteGroup
}

// Setup is used to setup the ghost project
// with the surrealdb database and gin router 
// engine. Template files are loaded from the
// src/views directory and static files are loaded
// from the static directory.
// 
// Example: 
//  ghostConfig, err := ghostutils.New() 
//  if err != nil { 
//      log.Fatal(err) 
//  } 
//  r := gin.Default() 
//  db, err := ghostConfig.Setup(r)
//  if err != nil {
//      log.Fatal(err)
//  }
//  r.Run(fmt.Sprintf(":%d", ghostConfig.Port))
// 
// Returns:
//  *surrealdb.DB for creating Routes using a GhostRoute interface 
//  error 
func (ghostConfig GhostConfig) Setup(r *gin.Engine) (*surrealdb.DB, error) {
    db, err := ghostConfig.surrealSetup()
    if err != nil {
        return db, err
    }
    r.LoadHTMLGlob("./src/views/**/*")
    r.Static("/static", "./static")
    return db, nil
}


func (ghostConfig GhostConfig) signinObj() map[string]interface{} {
    return map[string]interface{} {
        "user": ghostConfig.SurrealDB.Username,
        "pass": ghostConfig.SurrealDB.Password,
    }
}

func (ghostConfig GhostConfig) surrealSetup() (*surrealdb.DB, error) {
    var db *surrealdb.DB
    db, err := surrealdb.New(ghostConfig.SurrealDB.URL)
    if err != nil {
        return db, err
    }
    if _, err := db.Signin(
        ghostConfig.signinObj(),
    ) ; err != nil {
        return db, err
    }
    if _, err := db.Use(
        ghostConfig.SurrealDB.Namespace,
        ghostConfig.SurrealDB.Database,
    ); err != nil {
        return db, err
    }
    return db, nil
}
