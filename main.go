package main

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
	"time"
)

var sseClients = make(map[chan string]struct{})

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/interact", interactPage)
	http.HandleFunc("/interact/stream", interactStream)
	http.HandleFunc("/interact/process", interactProcess)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("Server starting on :%s\n", port)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	escaped := html.EscapeString(name)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Hello, %s!</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    }
    .card {
      background: #fff;
      border-radius: 16px;
      padding: 3rem 4rem;
      box-shadow: 0 20px 60px rgba(0,0,0,0.3);
      text-align: center;
    }
    h1 {
      font-size: 3rem;
      background: linear-gradient(135deg, #667eea, #764ba2);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
    p { color: #888; margin-top: 0.5rem; }
    a { color: #667eea; text-decoration: none; }
    a:hover { text-decoration: underline; }
    footer { margin-top: 1.5rem; font-size: 0.85rem; color: #bbb; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Hello, %s!</h1>
    <p>Your goTest server is running.</p>
    <footer>
      <a href="/health">health</a>
    </footer>
  </div>
</body>
</html>`, escaped, escaped)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		fmt.Fprintf(os.Stderr, "health: write failed: %v\n", err)
	}
}

func interactPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SSE Interactive</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      padding: 1rem;
    }
    .card {
      background: #fff;
      border-radius: 16px;
      padding: 2.5rem;
      box-shadow: 0 20px 60px rgba(0,0,0,0.3);
      width: 100%%;
      max-width: 600px;
    }
    h1 {
      font-size: 1.75rem;
      margin-bottom: 1.5rem;
      background: linear-gradient(135deg, #667eea, #764ba2);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
    .input-row {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 1rem;
    }
    input[type="text"] {
      flex: 1;
      padding: 0.75rem 1rem;
      border: 2px solid #e0e0e0;
      border-radius: 8px;
      font-size: 1rem;
      outline: none;
      transition: border-color 0.2s;
    }
    input[type="text"]:focus {
      border-color: #667eea;
    }
    button {
      padding: 0.75rem 1.5rem;
      background: linear-gradient(135deg, #667eea, #764ba2);
      color: #fff;
      border: none;
      border-radius: 8px;
      font-size: 1rem;
      cursor: pointer;
      font-weight: 600;
      transition: opacity 0.2s;
    }
    button:hover {
      opacity: 0.9;
    }
    button:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
    #output {
      background: #f5f5f5;
      border-radius: 8px;
      padding: 1rem;
      min-height: 200px;
      max-height: 400px;
      overflow-y: auto;
      font-family: "SF Mono", Monaco, monospace;
      font-size: 0.9rem;
      line-height: 1.6;
    }
    .msg {
      margin-bottom: 0.5rem;
      padding: 0.5rem;
      border-radius: 4px;
    }
    .msg.info { background: #e3f2fd; color: #1565c0; }
    .msg.success { background: #e8f5e9; color: #2e7d32; }
    .msg.error { background: #ffebee; color: #c62828; }
    footer {
      margin-top: 1.5rem;
      text-align: center;
      font-size: 0.85rem;
      color: #bbb;
    }
    footer a { color: #667eea; text-decoration: none; }
    footer a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="card">
    <h1>SSE Interactive Demo</h1>
    <div class="input-row">
      <input type="text" id="input" placeholder="输入内容..." autocomplete="off" />
      <button id="sendBtn" onclick="send()">发送</button>
    </div>
    <div id="output"><div class="msg info">等待输入...</div></div>
    <footer>
      <a href="/">← 返回首页</a>
    </footer>
  </div>
  <script>
    const output = document.getElementById("output");
    const sendBtn = document.getElementById("sendBtn");
    const input = document.getElementById("input");
    let es = null;

    function appendMsg(text, type) {
      const div = document.createElement("div");
      div.className = "msg " + type;
      div.textContent = text;
      output.appendChild(div);
      output.scrollTop = output.scrollHeight;
    }

    function connectSSE() {
      es = new EventSource("/interact/stream");
      es.onopen = function() { appendMsg("SSE 连接已建立", "info"); };
      es.onerror = function() { appendMsg("SSE 连接错误", "error"); };
      es.addEventListener("message", function(e) {
        appendMsg("服务端: " + e.data, "success");
        sendBtn.disabled = false;
        input.value = "";
        input.focus();
      });
    }

    function send() {
      const text = input.value.trim();
      if (!text) return;
      if (!es || es.readyState === EventSource.CLOSED) {
        appendMsg("SSE 未连接，请刷新页面", "error");
        return;
      }
      sendBtn.disabled = true;
      appendMsg("你: " + text, "info");
      fetch("/interact/process", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: text })
      }).catch(function() {
        appendMsg("发送失败", "error");
        sendBtn.disabled = false;
      });
    }

    input.addEventListener("keydown", function(e) {
      if (e.key === "Enter") send();
    });

    connectSSE();
  </script>
</body>
</html>`)
}

func interactStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 10)
	sseClients[ch] = struct{}{}
	defer func() {
		close(ch)
		delete(sseClients, ch)
	}()

	// Send a welcome message
	msg := "data: 已就绪，请在输入框输入内容\n\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		return
	}
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case m, ok := <-ch:
			if !ok {
				return
			}
			msg := "data: " + m + "\n\n"
			if _, err := w.Write([]byte(msg)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func interactProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Read raw body for JSON
	buf := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(buf); err != nil && err.Error() != "EOF" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	msg := strings.TrimSpace(string(buf))
	// strip JSON wrapper if present
	if strings.HasPrefix(msg, `{"message":`) {
		parts := strings.SplitN(msg, `:"`, 2)
		if len(parts) == 2 {
			msg = strings.TrimSuffix(parts[1], `"}`)
		}
	}

	go func(text string) {
		responses := []string{
			fmt.Sprintf("收到: %s", text),
			fmt.Sprintf("长度: %d 字符", len([]rune(text))),
			fmt.Sprintf("反转: %s", reverse(text)),
			fmt.Sprintf("大写: %s", strings.ToUpper(text)),
			"处理完成!",
		}
		for _, resp := range responses {
			broadcastSSE(resp)
			time.Sleep(300 * time.Millisecond)
		}
	}(msg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, `{"status":"queued"}`)
}

func broadcastSSE(msg string) {
	for ch := range sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
