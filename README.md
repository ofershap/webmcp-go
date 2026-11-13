# webmcp-go

Expose your Go web app to in-browser AI agents, without writing the WebMCP plumbing by hand.

[WebMCP](https://github.com/webmachinelearning/webmcp) lets a web page register **tools** that browser AI agents can discover and call. webmcp-go gives you the server half of that loop in Go: declare tools with typed input schemas and handlers, and the package serves the registration script plus the invoke endpoint. An agent calling your tool lands in your Go handler; the handler's reply goes back to the agent.

## Compatibility, honestly

WebMCP is early. As of November 2026, per the [official implementation status](https://github.com/webmachinelearning/webmcp/blob/main/implementation-status.md):

- **Chrome 149+**: available via Origin Trial (or a chrome:// flag)
- **Edge 150+**: Origin Trial
- **ChatGPT Desktop**: supported
- **Firefox and Safari**: positions still open

So this is a first-mover library, not something your stable-channel users' browsers have today. If you want your Go app ready for when agents start arriving in Chrome, this is the wedge.

## Install

```sh
go get github.com/ofershap/webmcp-go
```

## Usage

```go
mux := webmcp.NewMux()

mux.Register(webmcp.Tool{
    Name:        "filter-templates",
    Description: "Filters templates by a natural-language visual description.",
    InputSchema: webmcp.ObjectSchema(
        webmcp.String("description", "Visual description of templates to show.").Req(),
    ),
    Handler: func(input json.RawMessage) (any, error) {
        var in struct {
            Description string `json:"description"`
        }
        if err := json.Unmarshal(input, &in); err != nil {
            return nil, err
        }
        matches := filterTemplatesInUI(in.Description)
        return matches, nil
    },
})

http.Handle("/.well-known/webmcp/", http.StripPrefix("", mux.ServeHTTP("/.well-known/webmcp")))
```

Then add one line to the pages where the tools should be discoverable:

```html
<script src="/.well-known/webmcp/register.js" defer></script>
```

That script registers every tool with `navigator.modelContext`, and each tool's `execute()` POSTs the agent's input back to your Go handler.

## What you get

- **Typed schemas from Go** - `ObjectSchema`, `String`, `Integer`, `Enum`, etc. produce the JSON Schema the agent sees.
- **One endpoint pair** - `GET {prefix}/register.js` and `POST {prefix}/invoke/{name}`. Mount on any `net/http` mux at any prefix.
- **Stdlib only** - no dependencies.
- **Graceful no-op** - in browsers without WebMCP, the registration script silently does nothing.

## Notes

- Tool names must match `[a-z0-9-]`.
- The registration script is generated from your tool list, so adding a tool in Go automatically exposes it to agents.
- Human-in-the-loop is on you: if a tool mutates state (checkout, publish, delete), confirm with the user inside your handler before committing. WebMCP agents will call what you register.

## License

MIT
