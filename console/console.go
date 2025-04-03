package main

import (
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
)

var currentDir string
var mu sync.Mutex

func main() {
    // Initialize the current working directory
    var err error
    currentDir, err = os.Getwd()
    if err != nil {
        fmt.Println("Error getting current directory:", err)
        return
    }

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/execute", executeHandler)

    fmt.Println("Starting server on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Error starting server:", err)
    }
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Command Executor</title>
        <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/jquery.terminal/2.35.0/css/jquery.terminal.min.css">
        <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.4/jquery.min.js"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery.terminal/2.35.0/js/jquery.terminal.min.js"></script>
        <style>
            body {
                font-family: Arial, sans-serif;
                background-color: #1e1e1e;
                color: #ffffff;
                margin: 0;
                padding: 0;
            }
            #terminal {
                width: 100%;
                height: 100vh;
            }
        </style>
    </head>
    <body>
        <div id="terminal"></div>
        <script>
            $(function() {
                $('#terminal').terminal(function(command, term) {
                    if (command.trim() === '') return;
                    term.pause();
                    $.post('/execute', { command: command }, function(response) {
                        term.echo(response);
                        term.resume();
                    }).fail(function(xhr) {
                        term.error(xhr.responseText);
                        term.resume();
                    });
                }, {
                    greetings: 'Welcome to the Web Terminal\nType your commands below:',
                    prompt: '> ',
                    name: 'web_terminal',
                    height: '100%',
                    width: '100%',
                });
            });
        </script>
    </body>
    </html>
    `
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(tmpl))
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    command := r.FormValue("command")
    if command == "" {
        http.Error(w, "Command cannot be empty", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    // Execute the PowerShell command in the current working directory
    cmd := exec.Command("powershell", "-NoProfile", "-Command", command)
    cmd.Dir = currentDir
    output, err := cmd.CombinedOutput()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error executing command: %v\n%s", err, output), http.StatusInternalServerError)
        return
    }

    // Update the current working directory if the command changes it
    if filepath.IsAbs(command) {
        currentDir = command
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write(output)
}