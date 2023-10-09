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
func Load() (GhostConfig, error) {
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
func (ghostConfig GhostConfig) BasicSurrealSetup(r *gin.Engine) (*surrealdb.DB, error) {
    db, err := ghostConfig.surrealSetup()
    if err != nil {
        return db, err
    }
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
