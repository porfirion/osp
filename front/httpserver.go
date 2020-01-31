package front

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/porfirion/osp/processor"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type Server interface {
	Start()
}

var logger = log.New(os.Stdout, "HTTP: ", 0)

var indexTemplate = template.Must(template.New("index").Parse(indexTemplateSource))
//var indexTemplate = template.Must(template.ParseFiles("index.html"))

const processErrorCookieName = "process-error"
const previewImagesLimit = 10

type processRequest struct {
	Filename      string
	Width, Height int

	Label  string
	Left   int
	Top    int
	Right  int
	Bottom int
}

type indexModel struct {
	Filename     string
	Errors       []string
	Previews     []string
	PreviewLeft  int
	PreviewRight int
	TotalFiles   int
}

func (m *indexModel) addError(err string) {
	m.Errors = append(m.Errors, err)
}

func addProcessErrorAndRedirect(w http.ResponseWriter, r *http.Request, errorText string, url string) {
	http.SetCookie(w, &http.Cookie{
		Name:  processErrorCookieName,
		Value: errorText,
		Path:  "/",
	})
	http.Redirect(w, r, url, 302)
}

func getProcessError(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(processErrorCookieName); err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:   processErrorCookieName,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		return c.Value, nil
	} else {
		return "", err
	}
}

type server struct {
	addr       string
	imgPath    string
	processor  processor.Processor
	httpServer *http.Server
	router     *mux.Router
}

func findCurrentIndex(selectedFile string, files []string) int {
	if selectedFile == "" || len(files) == 0 {
		return -1
	}

	for ind, f := range files {
		if f == selectedFile {
			return ind
		}
	}

	return -1
}

// takePreviews takes some filenames for previews from all available files, but not more than previewImagesLimit
// Additionally returns left and right indexes of taken files (index starting from 1)
func takePreviews(previewImagesLimit int, files []string, currentIndex int) (previews []string, l int, r int) {
	if len(files) == 0 || currentIndex < 0 || len(files) <= currentIndex {
		return nil, 0, 0
	}

	left := currentIndex - previewImagesLimit/2
	if left < 0 {
		left = 0
	}

	right := left + previewImagesLimit
	if right > len(files) {
		// we are new right bound
		right = len(files)

		// if we have enough files to fulfill all previews - let's do it
		if right-previewImagesLimit >= 0 {
			left = right - previewImagesLimit
		}
	}

	//logger.Printf("%d files  left: %d right: %d", len(files), left, right)

	res := make([]string, 0, right-left)

	for _, f := range files[left:right] {
		res = append(res, f)
	}

	return res, left + 1, right
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	logger.Println("index request")

	model := &indexModel{}

	// if file specified in params, lets find it
	if filenames, ok := r.URL.Query()["filename"]; ok && len(filenames) > 0 {
		filename := filenames[0]
		filepath := path.Join(s.imgPath, filename)
		if info, err := os.Stat(filepath); err == nil && !info.IsDir() {
			// file exists
			model.Filename = filename
		} else {
			// file doesn't exist
			model.addError(fmt.Sprintf(`file "%s" doesn't exist`, filename))
		}
	}

	// Let's find previews in imgPath
	foundFiles, err := ioutil.ReadDir(s.imgPath)
	if err != nil {
		// should show that we have problems with directory
		model.addError(fmt.Sprintf("error searching files: %v", err))
	} else if len(foundFiles) > 0 {
		files := make([]string, 0, len(foundFiles))
		for _, f := range foundFiles {
			if !f.IsDir() {
				files = append(files, f.Name())
			}
		}

		if len(files) > 0 {
			var currentInd = -1
			if model.Filename == "" {
				// first file will be current
				model.Filename = files[0]
				currentInd = 0
			} else {
				currentInd = findCurrentIndex(model.Filename, files)
			}

			model.Previews, model.PreviewLeft, model.PreviewRight = takePreviews(previewImagesLimit, files, currentInd)
			model.TotalFiles = len(files)
		} else {
			// no files found (assuming there were some directories)
		}
	} else {
		//imgPath is empty.. nothing to do
	}

	if v, err := getProcessError(w, r); err == nil {
		model.addError(v)
	}

	if err := indexTemplate.Execute(w, model); err != nil {
		logger.Printf("error rendering template: %v\n", err)
		http.Error(w, "error rendering template", http.StatusInternalServerError)
	}
}

func (s *server) processHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("process request\n")

	if err := r.ParseForm(); err != nil {
		logger.Printf("error parsing request")
		addProcessErrorAndRedirect(w, r, "error parsing request", r.Referer())
		return
	}

	req := &processRequest{}

	decoder := schema.NewDecoder()
	if err := decoder.Decode(req, r.PostForm); err != nil {
		logger.Printf("error parsing form: %v", err)
		addProcessErrorAndRedirect(w, r, "error parsing response", r.Referer())
		return
	}

	req.Filename = strings.Trim(req.Filename, " \n")
	if req.Filename == "" {
		logger.Print("filename not specified")
		addProcessErrorAndRedirect(w, r, "Filename not specified", r.Referer())
		return
	}

	req.Label = strings.Trim(req.Label, " \n")
	if req.Label == "" {
		logger.Printf("label is empty")
		addProcessErrorAndRedirect(w, r, "Label is empty", "/?filename="+req.Filename)
		return
	}

	if req.Right == req.Left || req.Bottom == req.Top {
		logger.Printf("area has zero size")
		addProcessErrorAndRedirect(w, r, "Area has zero size", "/?filename="+req.Filename)
		return
	}

	logger.Printf("processing file %v\n", req)

	resp, err := s.processor.ProcessImage(req.Filename, req.Width, req.Height, req.Label, req.Left, req.Top, req.Right, req.Bottom)
	if err != nil {
		addProcessErrorAndRedirect(w, r, fmt.Sprintf("error sending request to processor: %v", err), "/?filename="+req.Filename)
		return
	}

	logger.Printf("processor response: %#v\n", resp)

	http.Redirect(w, r, "/", 302)
}

func (s *server) Start() {
	s.router = mux.NewRouter()

	s.router.PathPrefix("/img/").Handler(http.StripPrefix("/img", http.FileServer(http.Dir(s.imgPath))))

	s.router.HandleFunc("/", s.indexHandler)
	s.router.HandleFunc("/process", s.processHandler)

	logger.Printf("starting web server on %s", s.addr)

	s.httpServer = &http.Server{
		Handler: s.router,
		Addr:    s.addr,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			logger.Fatal("error starting web server", err)
		}
	}()
}

// StartServer starts new http server on specified host and port
func NewServer(host, port string, imgPath string, processor processor.Processor) (Server, error) {
	srv := &server{
		processor: processor,
		imgPath:   imgPath,
		addr:      host + ":" + port,
	}

	return srv, nil
}
