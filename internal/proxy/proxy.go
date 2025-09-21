package proxy

import (
	"caching-proxy/internal/cache"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/url"
)

type HandlerProxy struct {
	TargetUrl string
	Cache     cache.Cache
}

// ServeHTTP - главная функция-обработчик, которую вызывает Go при каждом запросе
// Решает: брать из кэша или перенаправить на целевой сервер
func (h *HandlerProxy) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET"{
		h.proxyWithoutCache(w, r)
		return
	}

	cacheKey := h.generateKey(r.URL.String())

	ctx := context.Background()
	if data, err := h.Cache.GetCache(ctx, cacheKey); err == nil && data != nil{
		log.Printf("CACHE HIT: %s", r.URL.Path)
		w.Header().Set("X-Cache", "HIT")
		w.Write(data)
		return
	}

	log.Printf("CACHE MISS: %s", r.URL.Path)
	h.proxyAndCache(w, r, cacheKey)
}

// generateKey - создает уникальный ключ для кэша на основе URL
// Нужна чтобы превратить длинные URL в короткие ключи для Redis
func (h HandlerProxy) generateKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// proxyAndCache - перенаправляет запрос на целевой сервер и сохраняет ответ в кэш
// Вызывается когда в кэше нет данных (CACHE MISS)
func (h *HandlerProxy) proxyAndCache(w http.ResponseWriter, r *http.Request, cacheKey string){
	targetUrl, err := url.Parse(h.TargetUrl + r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	resp, err := http.Get(targetUrl.String())
	if err != nil {
		http.Error(w, "Target server error", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) //Почему в одной функуции мы пропускаем err а в другой нет
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode == http.StatusOK {
		if err := h.Cache.SetCache(ctx, cacheKey, body); err != nil {
			log.Printf("Cache set error: %v", err)
		}
	}

	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(resp.StatusCode)
	w.Write(body) 
}


// proxyWithoutCache - просто перенаправляет запрос без кэширования
// Используется для POST, PUT, DELETE запросов (которые не нужно кэшировать)
func (h HandlerProxy) proxyWithoutCache(w http.ResponseWriter, r *http.Request) {
	targetUrl, err := url.Parse(h.TargetUrl + r.URL.Path)
	if err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	resp, err := http.Get(targetUrl.String())
	if err != nil {
		http.Error(w, "Target server error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}