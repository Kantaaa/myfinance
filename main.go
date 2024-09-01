package main

import (
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"log"
	"mymodule/pkg/taxCalculation"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Command-line flags.
var (
	httpAddr   = flag.String("addr", "localhost:8080", "Listen address")
	pollPeriod = flag.Duration("poll", 5*time.Second, "Poll period")
	version    = flag.String("version", "1.20", "Go version")
)

const baseChangeURL = "https://go.googlesource.com/go/+/"

func main() {
	flag.Parse()
	changeURL := fmt.Sprintf("%sgo%s", baseChangeURL, *version)
	http.Handle("/", NewServer(*version, changeURL, *pollPeriod))
	log.Printf("serving http://%s", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
	debugTaxCalculation()

}

// Exported variables for monitoring the server.
// These are exported via HTTP as a JSON object at /debug/vars.
var (
	hitCount       = expvar.NewInt("hitCount")
	pollCount      = expvar.NewInt("pollCount")
	pollError      = expvar.NewString("pollError")
	pollErrorCount = expvar.NewInt("pollErrorCount")
)

// Server implements the outyet server.
// It serves the user interface (it's an http.Handler)
// and polls the remote repository for changes.
type Server struct {
	version string
	url     string
	period  time.Duration

	mu  sync.RWMutex // protects the yes variable
	yes bool
}

// NewServer returns an initialized outyet server.
func NewServer(version, url string, period time.Duration) *Server {
	s := &Server{version: version, url: url, period: period}
	go s.poll()
	return s
}

// poll polls the change URL for the specified period until the tag exists.
// Then it sets the Server's yes field true and exits.
func (s *Server) poll() {
	for !isTagged(s.url) {
		pollSleep(s.period)
	}
	s.mu.Lock()
	s.yes = true
	s.mu.Unlock()
	pollDone()
}

// Hooks that may be overridden for integration tests.
var (
	pollSleep = time.Sleep
	pollDone  = func() {}
)

// isTagged makes an HTTP HEAD request to the given URL and reports whether it
// returned a 200 OK response.
func isTagged(url string) bool {
	pollCount.Add(1)
	r, err := http.Head(url)
	if err != nil {
		log.Print(err)
		pollError.Set(err.Error())
		pollErrorCount.Add(1)
		return false
	}
	return r.StatusCode == http.StatusOK
}

func formatNumber(n float64) string {
	return strings.Replace(fmt.Sprintf("%0.f", n), ",", " ", -1)
}

// ServeHTTP implements the HTTP user interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hitCount.Add(1)
	s.mu.RLock()
	data := struct {
		URL           string
		Version       string
		Yes           bool
		TaxCalculated bool
		Income        string
		Tax           string
		EffectiveTax  string
	}{
		URL:     s.url,
		Version: s.version,
		Yes:     s.yes,
	}
	s.mu.RUnlock()

	if r.Method == http.MethodPost {
		r.ParseForm()
		incomeStr := r.FormValue("income")
		income, err := strconv.ParseFloat(incomeStr, 64)
		if err == nil {
			data.Income = formatNumber(income) + " kr"
			tax := taxCalculation.IncomeTax(income)
			data.Tax = formatNumber(tax) + " kr"
			effectiveTax := taxCalculation.EffectiveTax(income)
			data.EffectiveTax = fmt.Sprintf("%.2f%%", effectiveTax)
			data.TaxCalculated = true
		} else {
			log.Print("Invalid income input: ", err)
		}
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Print(err)
	}
	/*
		<h2>Income Tax Calculator</h2>
		<form method="POST" action="/">
			<label for="income">Enter your income:</label>
			<input type="number" id="income" name="income" step="100000" required>
			<input type="submit" value="Calculate Tax">
		</form>
		{{if .TaxCalculated}}
			<h3>For an income of {{.Income}}, the calculated tax is {{.Tax}}</h3>
			<h3>The effective tax rate is {{.EffectiveTax}}</h3>
		{{end}}
	*/
}

func debugTaxCalculation() {
	income := 500000.0
	tax := taxCalculation.IncomeTax(income)
	fmt.Printf("Income: %f, Tax: %f\n", income, tax)
}

// tmpl is the HTML template that drives the user interface.
var tmpl = template.Must(template.New("tmpl").Parse(`
<!DOCTYPE html><html><body><center>
	<H1>Norsk skattekalkulator</H1>
	<form method="POST" action="/">	
	<label for="income"> Legg inn årlig bruttoinntekten: </label>
	<input type="float64" id="income" name="income" step="100000" required>
	<input type="submit" value="Regn ut!">
	</form>

	{{if .TaxCalculated}}
			<p>Ved bruttoinntekt av {{.Income}}, er din skatt følgende, {{.Tax}}</p>
			<p>Effektiv skattesats {{.EffectiveTax}}</p>
			
		{{end}}
</center>
</body></html>
`))
