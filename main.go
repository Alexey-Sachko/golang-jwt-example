package main

// Импортируем необходимые зависимости. Мы будем использовать
// пакет из стандартной библиотеки и пакет от gorilla
import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
)

// Product тип продуска
type Product struct {
	ID int
	Name string
}

var products = []Product{
	Product{0, "Ogurets"},
	Product{1, "Pomidor"},
	Product{2, "Banan"},
}

func main() {
	r := mux.NewRouter()
	// Наше API состоит из трех роутов
	// /status - нужен для проверки работоспособности нашего API
	// /products - возвращаем набор продуктов, 
	// по которым мы можем оставить отзыв
	// /products/{slug}/feedback - отображает фидбек пользователя по продукту
	r.Handle("/status", StatusHandler).Methods("GET")
	r.Handle("/products", jwtMiddleware.Handler(ProductsHandler)).Methods("GET")
	r.Handle("/products", jwtMiddleware.Handler(AddProductHandler)).Methods("POST")
	r.Handle("/logon", GetTokenHandler).Methods("GET")

	// Страница по умолчанию для нашего сайта это простой html.
	r.Handle("/", http.FileServer(http.Dir("./views/")))
	
	// Статику (картинки, скрипти, стили) будем раздавать 
	// по определенному роуту /static/{file} 
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
															http.FileServer(http.Dir("./static/"))))
	
	// Наше приложение запускается на 3000 порту. 
	// Для запуска мы указываем порт и наш роутер
	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, r))
}

// NotImplemented Этот хендлер просто возвращает сообщение "Not Implemented"
var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
  w.Write([]byte("Not Implemented"))
})

// StatusHandler Возвращает статус апи
var StatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("API is up and running"))
})

//ProductsHandler Этот хендлер возвращает пользователю список продуктов для оценки.
var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
	payload, _ := json.Marshal(products)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
})

// AddProductHandler - добавить продукт
var AddProductHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
  var product Product
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &product)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	product.ID = products[len(products) - 1].ID + 1

	products = append(products, product)

	w.WriteHeader(http.StatusAccepted)
})

// Глобальный секретный ключ
var mySigningKey = []byte("secret")

var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Создаем новый токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"admin": true,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		})

		// Подписываем токен нашим секретным ключем
		tokenString, _ := token.SignedString(mySigningKey)

		// Отдаем токен клиенту
		w.Write([]byte(tokenString))
})

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})