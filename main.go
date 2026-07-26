// تطبيق جمع بيانات UX Discovery — خادم واحد يخدم الواجهة والـ API معًا.
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"surveyapp/internal/api"
	"surveyapp/internal/seed"
	"surveyapp/internal/store"
)

//go:embed web
var webFS embed.FS

func main() {
	log.SetFlags(log.LstdFlags)

	// نفس البايناري يخدم فحص الصحة داخل الحاوية، فلا نحتاج curl في الصورة.
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		healthcheck()
		return
	}

	addr := env("ADDR", ":8080")
	dbPath := env("DB_PATH", "data/survey.db")
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		log.Fatal("المتغيّر ADMIN_PASSWORD مطلوب — عيّنه في ملف .env أو في docker-compose.yml")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("إنشاء مجلد البيانات: %v", err)
	}

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("فتح قاعدة البيانات: %v", err)
	}
	defer st.Close()

	n, err := seed.Apply(st)
	if err != nil {
		log.Fatalf("تحميل الأسئلة الأولية: %v", err)
	}
	if n > 0 {
		log.Printf("حُمِّل %d سؤالًا أوليًا", n)
	}

	web, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("تحميل ملفات الواجهة: %v", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.New(st, web, dbPath, password).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("الخادم يعمل على %s (قاعدة البيانات: %s)", addr, dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("تشغيل الخادم: %v", err)
		}
	}()

	// إيقاف مرتّب حتى لا تُقطع كتابة جارية على قاعدة البيانات.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("جارٍ الإيقاف...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("إيقاف الخادم: %v", err)
	}
}

// healthcheck يستعلم عن /healthz محليًا ويخرج بالرمز المناسب.
func healthcheck() {
	addr := env("ADDR", ":8080")
	if addr[0] == ':' {
		addr = "127.0.0.1" + addr
	}
	res, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		log.Printf("فحص الصحة: %v", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
