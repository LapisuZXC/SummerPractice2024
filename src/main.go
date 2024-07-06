package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"net/http"

	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

type Data struct {
	Items    []interface{}`json:"items"`
	Found    int      `json:"found"`
	Page     int      `json:"page"`
	Pages    int      `json:"pages"`
	Per_page int8      `json:"per_page"`
}

func main() {
	connStr := "postgres://postgres:secret@localhost:5432/gopgtest?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		if err = db.Ping(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Successfully connected!")

		createVacancyTable(db)

		
	
	e := echo.New()
	e.GET("/", func(c echo.Context) error {

		return c.File("html.html")
	})

	type SearchRequest struct {
		Name string `json:"name"`
		Clarification string `json:"clarification"`
		Salary string `json:"salary"`
		Location string `json:"location"`
	}
	e.POST("/search", func(c echo.Context) error {

		req := SearchRequest{}
		fmt.Printf("request: %v\n", req)
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		fmt.Println(req)
	
		return c.JSON(http.StatusOK, callToDB(db, req.Name))
	})
	e.Logger.Fatal(e.Start(":1323"))
	
	

}
func callToDB(db *sql.DB, str string) map[string]interface{} {
	reqStr := fmt.Sprintf("https://api.hh.ru/vacancies?text=%s&per_page=15", str)
	req, err := http.NewRequest("GET", reqStr, nil)
if err != nil {
	log.Fatal(err)
}



req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
client := &http.Client{}
resp, err := client.Do(req)
fmt.Println(resp)
fmt.Println(resp.StatusCode)
if err != nil {
	log.Fatal(err)
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
if err != nil {
	log.Fatal(err)
}
var data Data
json.Unmarshal(body, &data)
pk := insertVacancy(db, data)


var max int
query := `SELECT * FROM vacancy WHERE id = $1`
length  := db.QueryRow(`SELECT MAX(id) FROM vacancy`).Scan(&max)
if length != nil {
	log.Fatal(length)
}
var name string
var salary string
var area string
var url string
var l = make(map[string]interface{})
for i := 1; i < max+1; i++ {
	err = db.QueryRow(query, i).Scan(&pk, &name, &salary, &area, &url)
	if err != nil {
		log.Fatal(err)
	}
	l[strconv.FormatInt(int64(i),8)] = map[string]interface{}{"name": name, "salary": salary, "area": area, "url": url}
	if err != nil {
		   log.Fatal(err)
	}
	
}
fmt.Println(l)
return l
}

func createVacancyTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS "vacancy" (
		"id" SERIAL PRIMARY KEY,
		"name" VARCHAR(255) NOT NULL,
		"salary" VARCHAR(255) NOT NULL,
		"area" VARCHAR(255) NOT NULL,
		"vacancy_url" VARCHAR(255) NOT NULL
		);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertVacancy(db *sql.DB, data Data) int {
	db.Exec(`TRUNCATE TABLE "vacancy" CASCADE;
	ALTER SEQUENCE vacancy_id_seq RESTART WITH 1`)
	var pk int
	for _, arr := range data.Items {
	 if item, ok := arr.(map[string]interface{}); ok {
	  name, ok := item["name"].(string)
	  if !ok {
	   log.Println("Error converting name to string")
	   continue
	  }
   
	  salary := item["salary"]
   var salaryFrom, salaryTo, salaryCurrency string
   switch salary.(type) {
   case nil:
    // Handle the case where salary is null
    salaryFrom = ""
    salaryTo = ""
    salaryCurrency = ""
   case map[string]interface{}:
    salaryMap := salary.(map[string]interface{})

    if salaryFromVal, ok := salaryMap["from"]; ok {
     switch salaryFromVal.(type) {
     case float64:
      salaryFrom = strconv.FormatFloat(salaryFromVal.(float64), 'f', -1, 64)
     case int:
      salaryFrom = strconv.Itoa(salaryFromVal.(int))
	 case nil:
		salaryTo = ""
     default:
      log.Println("Invalid type for salary from")
      continue
     }
    }

    if salaryToVal, ok := salaryMap["to"]; ok {
     switch salaryToVal.(type) {
     case float64:
      salaryTo = strconv.FormatFloat(salaryToVal.(float64), 'f', -1, 64)
     case int:
      salaryTo = strconv.Itoa(salaryToVal.(int))
     case nil:
      salaryTo = ""
     default:
      log.Println("Invalid type for salary to")
      continue
     }
    }

	if salaryCurrencyVal, ok := salaryMap["currency"]; ok{
	switch salaryMap["currency"].(type) { 
	case string:
		salaryCurrency = salaryCurrencyVal.(string)
	default:
		log.Println("Invalid type for salary currency")
		continue
	}
	}	
    if !ok {
     log.Println("Error converting salary currency to string")
     continue
    }
   
	  area, ok := item["area"].(map[string]interface{})
	  if !ok {
	   log.Println("Error converting area to map")
	   continue
	  }
   
	  areaName, ok := area["name"].(string)
	  if !ok {
	   log.Println("Error converting area name to string")
	   continue
	  }
   
	  alternate_url, ok := item["alternate_url"].(string)
	  if !ok {
	   log.Println("Error converting alternative URL to string")
	   continue
	  }
   
	  query := `INSERT INTO "vacancy" ("name", "salary", "area", "vacancy_url") VALUES ($1, $2, $3, $4) RETURNING id`
	  err := db.QueryRow(query, name, salaryFrom+"-"+salaryTo+salaryCurrency, areaName, alternate_url).Scan(&pk)
	  if err != nil {
	   log.Println(err)
	   continue
	  }
	 }
	}
	}		
	return pk
}
	

