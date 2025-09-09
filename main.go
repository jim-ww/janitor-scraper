package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.ngrok.com/ngrok/v2"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	serveLocally = flag.Bool("local", false, "serve locally, instead of using ngrok")
	logRequests  = flag.Bool("log-requests", false, "log requests")
	address      = flag.String("address", "localhost:8080", "address to listen on (used only if served locally)")
	format       = flag.String("format", "plain", "messages output format: plain/json")
	noColor      = flag.Bool("no-color", false, "disable colored output")
	filePath     = flag.String("filepath", "", "if specified, messages will be saved to this file. example: /path/to/file.json")
)

func main() {
	flag.Parse()

	r := chi.NewRouter()

	if *logRequests {
		r.Use(middleware.Logger)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.Post("/v1/chat/completions", handleChatCompletions)

	if *serveLocally {
		fmt.Println("Server is listening on: ", *address)
		log.Fatal(http.ListenAndServe(*address, r))
	} else {
		// else serve with ngrok
		if os.Getenv("NGROK_AUTHTOKEN") == "" {
			log.Fatal("NGROK_AUTHTOKEN environment variable is not set. Please set it to your ngrok authtoken. You can obtain an authtoken from https://dashboard.ngrok.com/get-started/your-authtoken")
		}

		l, err := ngrok.Listen(context.Background(), ngrok.WithURL(os.Getenv("NGROK_RESERVED_URL")))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Server is listening on: ", l.Addr())
		log.Fatal(http.Serve(l, r))
	}
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Messages []Message `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error decoding request body", "err", err)
		return
	}

	if len(req.Messages) == 0 {
		slog.Error("No messages found in request")
		return
	}

	wr := os.Stdout
	if *filePath != "" {
		if _, err := os.Stat(*filePath); err == nil {
			slog.Error("File already exists", "file-path", *filePath)
			return
		}
		file, err := os.Create(*filePath)
		if err != nil {
			slog.Error("Error creating file", "err", err)
			return
		}
		defer file.Close()
		wr = file
	}

	if *format == "json" {
		if err := json.NewEncoder(wr).Encode(req.Messages); err != nil {
			slog.Error("Error encoding messages to JSON", "err", err)
			return
		}
		fmt.Println("Messages written to file.")
		return
	}

	fmt.Fprintln(wr, "\n"+strings.Repeat("═", 80))
	fmt.Fprintln(wr, "💬 CHAT")
	fmt.Fprintln(wr, strings.Repeat("═", 80))

	for i, msg := range req.Messages {
		fmt.Fprintf(wr, "\n%s ", getRoleEmoji(msg.Role))
		colorEnabled := !*noColor && *filePath == ""
		printMessage(wr, msg, i, colorEnabled)
	}

	fmt.Println(strings.Repeat("═", 80) + "\n")
}

func printMessage(wr io.Writer, msg Message, index int, colorEnabled bool) {
	roleDisplay := strings.ToUpper(msg.Role)

	if colorEnabled {
		fmt.Fprintf(wr, "\x1b[1m%s\x1b[0m ", roleDisplay)
		fmt.Fprintf(wr, "(\x1b[90m%d\x1b[0m):\n", index+1)
	} else {
		fmt.Fprintf(wr, "%s (%d):\n", roleDisplay, index+1)
	}

	content := strings.TrimSpace(msg.Content)
	content = strings.ReplaceAll(content, "\\n", "\n")

	if colorEnabled {
		fmt.Fprintf(wr, "\x1b[97m%s\x1b[0m\n", formatText(content, 70))
	} else {
		fmt.Fprintf(wr, "%s\n", formatText(content, 70))
	}
}

func formatText(text string, maxWidth int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+len(word)+1 <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

func getRoleEmoji(role string) string {
	switch role {
	case "system":
		return "⚙️"
	case "user":
		return "👤"
	case "assistant":
		return "🤖"
	default:
		return "❓"
	}
}
