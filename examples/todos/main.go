// Command todos is a minimal webmcp-go demo: a todo page that registers
// add-todo and list-todos tools for in-browser agents.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/ofershap/webmcp-go"
)

var (
	mu    sync.Mutex
	todos []string
)

func main() {
	m := webmcp.NewMux()

	if err := m.Register(webmcp.Tool{
		Name:        "add-todo",
		Description: "Adds an item to the todo list on this page.",
		InputSchema: webmcp.ObjectSchema(
			webmcp.String("text", "The todo item text.").Req(),
		),
		Handler: func(input json.RawMessage) (any, error) {
			var in struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(input, &in); err != nil {
				return nil, err
			}
			mu.Lock()
			todos = append(todos, in.Text)
			mu.Unlock()
			return map[string]bool{"added": true}, nil
		},
	}); err != nil {
		log.Fatal(err)
	}

	if err := m.Register(webmcp.Tool{
		Name:        "list-todos",
		Description: "Lists the current todo items.",
		Handler: func(input json.RawMessage) (any, error) {
			mu.Lock()
			defer mu.Unlock()
			return todos, nil
		},
	}); err != nil {
		log.Fatal(err)
	}

	http.Handle("/webmcp/", m.ServeHTTP("/webmcp"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><title>webmcp-go demo</title>
<h1>Todos</h1>` + webmcp.ScriptTag("/webmcp") + `<p>Open this page in a WebMCP-capable browser and ask the agent to add a todo.</p>`))
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
