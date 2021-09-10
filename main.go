package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
)

func init() {
	fmt.Println("turboRappiMiddleware plugin is loaded!")
}

func main() {}

// HandlerRegisterer is the name of the symbol krakend looks up to try and register plugins
var HandlerRegisterer registrable = registrable("turboRappiMiddleware")

type registrable string

const outputHeaderName = "X-User-Data"
const pluginName = "turboRappiMiddleware"

func (r registrable) RegisterHandlers(f func(
	name string,
	handler func(
		context.Context,
		map[string]interface{},
		http.Handler) (http.Handler, error),
)) {
	f(pluginName, r.registerHandlers)
}

func (r registrable) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {

	hostUser := "http://localhost:3001/api/getUserData"
	client := &http.Client{Timeout: 3 * time.Second}

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var auth = r.Header.Get("Authorization")
		tokenString := strings.Fields(auth)[1]
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {

			fmt.Println(claims)
			email := claims["email"]

			val, err := rdb.Get(ctx, email).Result()
			if err != nil {
				panic(err)
			}
			fmt.Println(email, val)

			rq, err := http.NewRequest(http.MethodGet, hostUser, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rq.Header.Set("Content-Type", "application/json")

			rs, err := client.Do(rq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotAcceptable)
				return
			}
			defer rs.Body.Close()

			rsBodyBytes, err := ioutil.ReadAll(rs.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotAcceptable)
				return
			}

			r2 := new(http.Request)
			*r2 = *r

			r2.Header.Set(outputHeaderName, string(rsBodyBytes))

			handler.ServeHTTP(w, r2)
		} else {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
	}), nil
}
