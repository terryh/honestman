package main

import (
	"crypto/tls"
	"fmt"
	"honestman/app"
	"honestman/schema"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "net/http/pprof"

	"html/template"

	"github.com/NYTimes/gziphandler"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/russross/blackfriday"
	httpware "github.com/terryh/go-httpware"
	"github.com/unrolled/render"
	"golang.org/x/crypto/acme/autocert"
	//"github.com/davecheney/profile"
)

var (
	// AppContext hold share object
	AppContext *app.Context
	// Render ...
	Render           *render.Render
	templateForDoc   *template.Template
	templateForIndex *template.Template

	// clean up
	clean = func(s string) (qs []interface{}) {
		for _, w := range strings.Split(s, " ") {
			if strings.TrimSpace(w) != "" {
				qs = append(qs, strings.TrimSpace(w))
			}
		}
		return qs
	}
)

// Index for website
// FIXME, only for DEMO, or let's put handler to struct handler directory
func Index(w http.ResponseWriter, r *http.Request) {
	// Render.HTML(w, http.StatusOK, "index", map[string]string{})
	w.Header().Set("Content-type", "text/html;charset=utf-8")
	templateForIndex.Execute(w, nil)
}

//DocHandler document
func DocHandler(w http.ResponseWriter, r *http.Request) {
	readme := FSMustByte(AppContext.Debug, "/static/README.md")
	content := blackfriday.MarkdownCommon(readme)

	var doc = struct {
		Authors     string
		PackageName string
		Version     string
		LastUpdated string
		Content     template.HTML
	}{app.Authors, app.PackageName, app.Version, app.LastUpdated, template.HTML(content)}

	//templateForDoc.Execute(w, doc)
	w.Header().Set("Content-type", "text/html;charset=utf-8")
	templateForDoc.Execute(w, doc)
}

// APIHandler DEMO simple search
func APIHandler(w http.ResponseWriter, r *http.Request) {
	var db *sqlx.DB
	var err error
	var per_page, page, count int
	var items []schema.Item
	var where []string
	var ctx = make(map[string]interface{})
	per_page = 50
	page = 1

	db = r.Context().Value("db").(*sqlx.DB)

	pageStr := r.URL.Query().Get("page")

	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	ctx["page"] = page

	kwStr := r.URL.Query().Get("q")
	kwSlice := clean(kwStr)

	for idx := range kwSlice {
		where = append(where, fmt.Sprintf("name ~ $%d", idx+1))
	}
	log.Println(where)
	log.Println(kwSlice)
	// handle paging
	limitoffset := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(kwSlice)+1, len(kwSlice)+2)
	kwSlice = append(kwSlice, per_page)
	kwSlice = append(kwSlice, (page-1)*per_page)
	log.Println(limitoffset)

	// not include limit offset
	if len(kwSlice) > 2 {

		err = db.Get(&count, fmt.Sprintf("SELECT count(*) as count FROM item  WHERE %s", strings.Join(where, " AND ")), kwSlice[:len(kwSlice)-2]...)
		if err != nil {
			log.Println(err)
			ctx["error"] = err.Error()
		} else {
			ctx["count"] = count
		}

		err = db.Select(&items, fmt.Sprintf("SELECT * FROM item  WHERE %s ORDER BY price %s", strings.Join(where, " AND "), limitoffset), kwSlice...)
		if err != nil {
			log.Println(err)
			ctx["error"] = err.Error()
		} else {
			ctx["item"] = items
		}
	}

	Render.JSON(w, http.StatusOK, ctx)
}

// Main main
func Main(context *app.Context) *bone.Mux {

	Render = render.New(render.Options{
		Directory:  "/static",
		Extensions: []string{".html"},
		Asset: func(name string) ([]byte, error) {
			return FSByte(context.Debug, name)
		},
		AssetNames: func() []string {
			return []string{
				"/static/index.html",
			}
		},
	})

	templateForDoc, _ = template.New("doc").Parse(FSMustString(context.Debug, "/static/doc.html"))
	templateForIndex, _ = template.New("doc").Parse(FSMustString(context.Debug, "/static/index.html"))

	common := alice.New(
		httpware.SimpleLogger,
		httpware.Recovery,
		gziphandler.GzipHandler,
		cors.New(cors.Options{AllowedHeaders: []string{"*"}, AllowCredentials: true}).Handler,
		httpware.PostgresDB(context.DB, "db"),
	)

	mux := bone.New()

	// doc
	mux.Get("/doc", http.HandlerFunc(DocHandler))

	// index
	mux.Get("/", http.HandlerFunc(Index))

	// static files && static pages
	mux.Get("/static/*", http.FileServer(FS(context.Debug)))

	// api
	mux.Get("/api/search", common.ThenFunc(APIHandler))
	return mux
}

func main() {
	// init app context
	AppContext = app.NewContext()
	mux := Main(AppContext)

	// in production mode prepare https
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("in28.net"),
		Cache:      autocert.DirCache("certs"), //folder for storing certificates
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	if !AppContext.Debug {

		go func() {
			log.Printf("Starting HTTPS service on :443 ...")
			server.ListenAndServeTLS("", "") //key and cert are comming from Let's Encrypt
		}()
	}

	log.Printf("Starting HTTP service on %s ...", AppContext.Port)
	http.ListenAndServe(AppContext.Port, certManager.HTTPHandler(mux))
}
