package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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

	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("💬 CHAT")
	fmt.Println(strings.Repeat("═", 80))

	for i, msg := range req.Messages {
		fmt.Printf("\n%s. ", getRoleEmoji(msg.Role))
		printMessage(msg, i)
	}

	fmt.Println(strings.Repeat("═", 80) + "\n")
}

func printMessage(msg Message, index int) {
	roleDisplay := strings.ToUpper(msg.Role)

	fmt.Printf("\x1b[1m%s\x1b[0m ", roleDisplay)
	fmt.Printf("(\x1b[90m%d\x1b[0m):\n", index+1)

	content := strings.TrimSpace(msg.Content)
	content = strings.ReplaceAll(content, "\\n", "\n")

	fmt.Printf("\x1b[97m%s\x1b[0m\n", formatText(content, 70))
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
