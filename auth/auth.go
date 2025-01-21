package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
)

// validateToken(tokenString) проверяет правильность переданного JWT-токена
// и возвращает результат проверки вместе с возможной ошибкой.
func validateToken(tokenString string) (bool, error) {
	storedPassword := os.Getenv("TODO_PASSWORD")

	// Парсим токен и проверяем его с использованием секретного ключа
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(storedPassword), nil
	})

	if err != nil {
		// Возвращаем ошибку, если токен невалиден или его невозможно распарсить
		return false, err
	}

	// Проверяем валидность токена и его тип (jwt.MapClaims)
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	}

	// Возвращаем ошибку, если токен недействителен
	return false, fmt.Errorf("недействительный токен")
}

// Auth(next) создает middleware для проверки аутентификации перед обработкой запроса.
// Если токен недействителен или отсутствует, возвращается ошибка 401.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем секретный ключ из переменных окружения
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwtString string
			// Пытаемся получить токен из cookie
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtString = cookie.Value
			}

			// Проверяем валидность токена
			valid, err := validateToken(jwtString)
			if err != nil {
				valid = false
				fmt.Println("Не валидный токен: ", err)
			}

			// Если токен недействителен, возвращаем ошибку аутентификации 401
			if !valid {
				// Логирование ошибки аутентификации
				fmt.Println("Ошибка аутентификации: ", err)
				http.Error(w, "Аутентификация требуется", http.StatusUnauthorized)
				return
			}
		}
		// Если токен валиден, передаем управление следующему обработчику
		next(w, r)
	})
}
