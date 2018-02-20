package main

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs"
	"github.com/boutros/ulvemelk/data"
	"github.com/boutros/ulvemelk/data/template"
	"github.com/go-chi/chi"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var matcher = language.NewMatcher([]language.Tag{language.English, language.Norwegian})

type server struct {
	mux chi.Router
	sm  *scs.Manager
}

func newServer() *server {
	return &server{
		sm: scs.NewCookieManager("TODO"),
	}
}

func (s *server) setupRouting() {
	s.mux = chi.NewRouter()
	s.mux.Use(s.withMessagePrinter)
	s.mux.Get("/lang", func(w http.ResponseWriter, r *http.Request) {
		session := s.sm.Load(r)
		lang, _ := session.GetString("lang")
		//accept := r.Header.Get("Accept-Language")
		//fallback := "en"
		var newLang string
		if lang == "no" {
			newLang = "en"
		} else if lang == "en" {
			newLang = "no"
		} else {
			// TODO
			newLang = "en"
		}
		if err := session.PutString(w, "lang", newLang); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var tmpl template.Home
		tmpl.Render(r.Context(), w)
	})
	fs := http.StripPrefix("/static", http.FileServer(data.Assets))
	s.mux.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

func (s *server) withMessagePrinter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := s.sm.Load(r)
		lang, _ := session.GetString("lang")
		accept := r.Header.Get("Accept-Language")
		fallback := "en"
		tag := message.MatchLanguage(lang, accept, fallback)
		p := message.NewPrinter(tag)
		ctx := context.WithValue(r.Context(), template.MessagePrinterKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
