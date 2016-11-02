package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/lookingcloudy/bitbuckethook/hook"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var (
	ip             = flag.String("ip", "0.0.0.0", "ip the webhook should serve hooks on")
	port           = flag.Int("port", 9000, "port the webhook should serve hooks on")
	verbose        = flag.Bool("verbose", false, "show verbose output")
	noPanic        = flag.Bool("nopanic", false, "do not panic if hooks cannot be loaded when webhook is not running in verbose mode")
	hooksFilePath  = flag.String("hooks", "hooks.json", "path to the json file containing defined hooks the webhook should serve")
	hooksURLPrefix = flag.String("urlprefix", "hooks", "url prefix to use for served hooks (protocol://yourserver:port/PREFIX/:hook-id)")
	secure         = flag.Bool("secure", false, "use HTTPS instead of HTTP")
	cert           = flag.String("cert", "cert.pem", "path to the HTTPS certificate pem file")
	key            = flag.String("key", "key.pem", "path to the HTTPS certificate private key pem file")
	debug          = flag.Bool("debug", false, "true outputs the body of json")

	//responseHeaders hook.ResponseHeaders

	hooks hook.Hooks
)

func main() {
	hooks = hook.Hooks{}

	flag.Parse()

	log.SetPrefix("[webhook] ")
	log.SetFlags(log.Ldate | log.Ltime)

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	log.Println("starting")

	// load and parse hooks
	log.Printf("attempting to load hooks from %s\n", *hooksFilePath)

	err := hooks.LoadFromFile(*hooksFilePath)

	if err != nil {
		if !*verbose && !*noPanic {
			log.SetOutput(os.Stdout)
			log.Fatalf("couldn't load any hooks from file! %+v\naborting webhook execution since the -verbose flag is set to false.\nIf, for some reason, you want webhook to start without the hooks, either use -verbose flag, or -nopanic", err)
		}

		log.Printf("couldn't load hooks from file! %+v\n", err)
	} else {
		seenHooksIds := make(map[string]bool)

		log.Printf("found %d hook(s) in file\n", len(hooks))

		for _, hook := range hooks {
			if seenHooksIds[hook.ID] == true {
				log.Fatalf("error: hook with the id %s has already been loaded!\nplease check your hooks file for duplicate hooks ids!\n", hook.ID)
			}
			seenHooksIds[hook.ID] = true
			log.Printf("\tloaded: %s\n", hook.ID)
		}
	}

	l := negroni.NewLogger()
	l.Logger = log.New(os.Stderr, "[webhook] ", log.Ldate|log.Ltime)

	negroniRecovery := &negroni.Recovery{
		Logger:     l.Logger,
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}

	n := negroni.New(negroniRecovery, l)

	router := mux.NewRouter()

	var hooksURL string

	if *hooksURLPrefix == "" {
		hooksURL = "/{id}"
	} else {
		hooksURL = "/" + *hooksURLPrefix + "/{id}"
	}

	router.HandleFunc(hooksURL, hookHandler)

	n.UseHandler(router)

	if *secure {
		log.Printf("serving hooks on https://%s:%d%s", *ip, *port, hooksURL)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", *ip, *port), *cert, *key, n))
	} else {
		log.Printf("serving hooks on http://%s:%d%s", *ip, *port, hooksURL)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *ip, *port), n))
	}

}

func hookHandler(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	if matchedHook := hooks.Match(id); matchedHook != nil {
		log.Printf("%s got matched\n", id)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			if *debug {
				log.Println(string(body))
			}

			log.Printf("error reading the request body. %+v\n", err)
		}

		// parse body
		var payload hook.BitPush

		contentType := r.Header.Get("Content-Type")

		if strings.Contains(contentType, "json") {
			decoder := json.NewDecoder(strings.NewReader(string(body)))
			decoder.UseNumber()

			err := decoder.Decode(&payload)

			if err != nil {
				if *debug {
					log.Println(string(body))
				}

				log.Printf("error parsing JSON payload %+v\n", err)
			}
		}

		var ok bool
		var matchedValue string

		if matchedHook.TriggerRule == nil {
			ok = true
		} else {

			ok, matchedValue = matchedHook.Evaluate(&payload)
		}

		if ok {
			log.Printf("%s hook triggered successfully\n", matchedHook.ID)

			go handleHook(matchedHook, matchedValue)
			fmt.Fprintf(w, matchedHook.ResponseMessage)

			return
		}

		// if none of the hooks got triggered
		log.Printf("%s got matched, but didn't get triggered because the trigger rules were not satisfied\n", matchedHook.ID)
		fmt.Fprintf(w, "Hook rules were not satisfied.")
		if *debug {
			log.Println(string(body))
		}

	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Hook not found.")
	}
}

func handleHook(h *hook.Hook, matchedValue string) string {
	log.Println("Matched Value: ", matchedValue)
	cmd := exec.Command(h.ExecuteCommand, matchedValue)
	cmd.Dir = h.CommandWorkingDirectory
	//cmd.Args = []string{matchedValue}

	log.Printf("executing %s (%s) with arguments %q using %s as cwd\n", h.ExecuteCommand, cmd.Path, cmd.Args, cmd.Dir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running command: %v\n", err)
	}

	log.Printf("command output: %s\n", out)

	log.Printf("finished handling %s\n", h.ID)

	return string(out)
}
