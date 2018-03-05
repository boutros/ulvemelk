package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs"
	"github.com/boutros/ulvemelk"
	"github.com/boutros/ulvemelk/data"
	"github.com/boutros/ulvemelk/data/template"
	"github.com/go-chi/chi"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var matcher = language.NewMatcher([]language.Tag{language.English, language.Norwegian})

var predefSearches = []struct {
	Title string
	Img   string
	Desc  string
	Query string
}{
	{
		Title: "Nye soppbøker",
		Desc:  "Hold deg oppdatert!",
		Img:   "/static/q1.jpg",
		Query: "sopp",
	},
	{
		Title: "Mykologiske klassikere",
		Desc:  "Fordyp deg i tidløs visdom.",
		Img:   "/static/q2.jpg",
		Query: "sopp",
	},
	{
		Title: "Slimpsopp",
		Desc:  "Finn litteratur om denne fascinerende gruppen organismer.",
		Img:   "/static/q3.jpg",
		Query: "slimsopp",
	},
	{
		Title: "Bøker som ikke handler om sopp",
		Desc:  "Et lite utvalg for dere andre ignorante plebeiere.",
		Img:   "/static/q4.jpg",
		Query: "*",
	},
}

type Server struct {
	mux chi.Router
	sm  *scs.Manager
}

func NewServer() *Server {
	s := Server{
		sm: scs.NewCookieManager("TODO"),
	}
	s.setupRouting()
	return &s
}

func (s *Server) Serve() {
	if err := http.ListenAndServe(":4321", s.mux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) setupRouting() {
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
		tmpl.Searches = predefSearches
		tmpl.Render(r.Context(), w)
	})
	s.mux.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		res, err := http.Get("https://sok.deichman.no/q?query=" + q)
		if err != nil {
			log.Println(err)
			http.Error(w, "sorry sorry", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()
		var esRes esSearchResults
		if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
			log.Println(err)
			http.Error(w, "sorry sorry", http.StatusInternalServerError)
			return
		}
		tmpl := template.Search{
			Results: esRes.Massage(),
		}
		tmpl.Render(r.Context(), w)
	})

	fs := http.StripPrefix("/static", http.FileServer(data.Assets))
	s.mux.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

func (s *Server) withMessagePrinter(next http.Handler) http.Handler {
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

type oneOreMore []string

func (o oneOreMore) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	var res []string
	if b[0] == '[' {
		return json.Unmarshal(b, &res)
	}
	var one string
	err := json.Unmarshal(b, &one)
	res = append(res, one)
	return err
}

type esSearchResults struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int     `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Summary       string     `json:"summary"`
				WorkTypeLabel string     `json:"workTypeLabel"`
				Subject       []string   `json:"subject"`
				Author        oneOreMore `json:"author"`
				Subjects      []struct {
					URI string `json:"uri"`
				} `json:"subjects"`
				DisplayLine2 string `json:"displayLine2"`
				URI          string `json:"uri"`
				HasWorkType  struct {
					URI string `json:"uri"`
				} `json:"hasWorkType"`
				DisplayLine1    string     `json:"displayLine1"`
				Agents          oneOreMore `json:"agents"`
				Nationality     oneOreMore `json:"nationality"`
				MainEntryName   string     `json:"mainEntryName"`
				MainTitle       string     `json:"mainTitle"`
				TotalNumItems   int        `json:"totalNumItems"`
				Audiences       []string   `json:"audiences"`
				PublicationYear string     `json:"publicationYear"`
				Dewey           oneOreMore `json:"dewey"`
				Contributors    []struct {
					Agent struct {
						URI       string `json:"uri"`
						BirthYear string `json:"birthYear"`
						Name      string `json:"name"`
					} `json:"agent"`
					MainEntry bool   `json:"mainEntry"`
					Role      string `json:"role"`
				} `json:"contributors"`
				FictionNonfiction string `json:"fictionNonfiction"`
			} `json:"_source"`
			InnerHits struct {
				Publications struct {
					Hits struct {
						Total    int `json:"total"`
						MaxScore int `json:"max_score"`
						Hits     []struct {
							Type    string `json:"_type"`
							ID      string `json:"_id"`
							Score   int    `json:"_score"`
							Routing string `json:"_routing"`
							Parent  string `json:"_parent"`
							Source  struct {
								PublishedBy       string     `json:"publishedBy"`
								Isbn              oneOreMore `json:"isbn"`
								Language          oneOreMore `json:"language"`
								Title             oneOreMore `json:"title"`
								AvailableBranches []string   `json:"availableBranches"`
								RecordID          string     `json:"recordId"`
								Kd                string     `json:"kd"`
								Photographer      string     `json:"photographer"`
								Image             string     `json:"image"`
								Languages         []string   `json:"languages"`
								Created           time.Time  `json:"created"`
								Author            oneOreMore `json:"author"`
								Mt                string     `json:"mt"`
								DisplayLine2      string     `json:"displayLine2"`
								URI               string     `json:"uri"`
								DisplayLine1      string     `json:"displayLine1"`
								Agents            oneOreMore `json:"agents"`
								NumItems          int        `json:"numItems"`
								Nationality       oneOreMore `json:"nationality"`
								MainEntryName     string     `json:"mainEntryName"`
								MainTitle         string     `json:"mainTitle"`
								Subtitle          string     `json:"subtitle"`
								PublicationYear   string     `json:"publicationYear"`
								Contributors      []struct {
									Agent struct {
										URI  string `json:"uri"`
										Name string `json:"name"`
									} `json:"agent"`
									Role      string `json:"role"`
									MainEntry bool   `json:"mainEntry,omitempty"`
								} `json:"contributors"`
								WorkURI      string   `json:"workUri"`
								Mediatype    string   `json:"mediatype"`
								HomeBranches []string `json:"homeBranches"`
							} `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"publications"`
			} `json:"inner_hits"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Facets struct {
			DocCount int `json:"doc_count"`
			Facets   struct {
				DocCount  int `json:"doc_count"`
				Audiences struct {
					DocCount  int `json:"doc_count"`
					Audiences struct {
						DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
						SumOtherDocCount        int `json:"sum_other_doc_count"`
						Buckets                 []struct {
							Key      string `json:"key"`
							DocCount int    `json:"doc_count"`
						} `json:"buckets"`
					} `json:"audiences"`
				} `json:"audiences"`
				PublicationFacets struct {
					DocCount int `json:"doc_count"`
					Formats  struct {
						DocCount int `json:"doc_count"`
						Formats  struct {
							DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
							SumOtherDocCount        int `json:"sum_other_doc_count"`
							Buckets                 []struct {
								Key      string `json:"key"`
								DocCount int    `json:"doc_count"`
								Parents  struct {
									Value int `json:"value"`
								} `json:"parents"`
							} `json:"buckets"`
						} `json:"formats"`
					} `json:"formats"`
					Languages struct {
						DocCount  int `json:"doc_count"`
						Languages struct {
							DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
							SumOtherDocCount        int `json:"sum_other_doc_count"`
							Buckets                 []struct {
								Key      string `json:"key"`
								DocCount int    `json:"doc_count"`
								Parents  struct {
									Value int `json:"value"`
								} `json:"parents"`
							} `json:"buckets"`
						} `json:"languages"`
					} `json:"languages"`
					Mediatype struct {
						DocCount  int `json:"doc_count"`
						Mediatype struct {
							DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
							SumOtherDocCount        int `json:"sum_other_doc_count"`
							Buckets                 []struct {
								Key      string `json:"key"`
								DocCount int    `json:"doc_count"`
								Parents  struct {
									Value int `json:"value"`
								} `json:"parents"`
							} `json:"buckets"`
						} `json:"mediatype"`
					} `json:"mediatype"`
					HomeBranches struct {
						DocCount     int `json:"doc_count"`
						HomeBranches struct {
							DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
							SumOtherDocCount        int `json:"sum_other_doc_count"`
							Buckets                 []struct {
								Key      string `json:"key"`
								DocCount int    `json:"doc_count"`
								Parents  struct {
									Value int `json:"value"`
								} `json:"parents"`
							} `json:"buckets"`
						} `json:"homeBranches"`
					} `json:"homeBranches"`
				} `json:"publication_facets"`
				FictionNonfiction struct {
					DocCount          int `json:"doc_count"`
					FictionNonfiction struct {
						DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
						SumOtherDocCount        int `json:"sum_other_doc_count"`
						Buckets                 []struct {
							Key      string `json:"key"`
							DocCount int    `json:"doc_count"`
						} `json:"buckets"`
					} `json:"fictionNonfiction"`
				} `json:"fictionNonfiction"`
			} `json:"facets"`
		} `json:"facets"`
	} `json:"aggregations"`
}

func (es esSearchResults) Massage() (res ulvemelk.SearchResults) {
	for _, work := range es.Hits.Hits {
		var hit struct {
			Title string
		}
		hit.Title = work.Source.MainTitle
		res.Hits = append(res.Hits, hit)
	}
	return res
}
