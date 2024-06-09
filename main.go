package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "sync"
)

const (
    rows = 3
    cols = 3
)

type Game struct {
    Board  [][]string `json:"board"`
    Turn   string     `json:"turn"`
    Winner string     `json:"winner"`
    Draw   bool       `json:"draw"`
    Lock   sync.Mutex
}

func InitializeBoard() [][]string {
    board := make([][]string, rows)
    for i := range board {
        board[i] = make([]string, cols)
        for j := range board[i] {
            board[i][j] = " " // Use space for empty cells
        }
    }
    return board
}

func PrintBoard(board [][]string) {
    for i := 0; i < rows; i++ {
        for j := 0; j < cols; j++ {
            fmt.Printf("%s", board[i][j]) // Print each cell with its current content
        }
        fmt.Println() // Move to the next row
    }
}

func MakeMove(board [][]string, row, col int, player string) bool {
    if row >= 0 && row < rows && col >= 0 && col < cols && board[row][col] == " " {
        board[row][col] = player // Set the player's symbol in the chosen cell
        return true
    }
    return false
}

func CheckWin(board [][]string, player string) bool {
    // Check rows
    for i := 0; i < rows; i++ {
        if board[i][0] == player && board[i][1] == player && board[i][2] == player {
            return true
        }
    }

    // Check columns
    for j := 0; j < cols; j++ {
        if board[0][j] == player && board[1][j] == player && board[2][j] == player {
            return true
        }
    }

    // Check diagonals
    if board[0][0] == player && board[1][1] == player && board[2][2] == player {
        return true
    }
    if board[0][2] == player && board[1][1] == player && board[2][0] == player {
        return true
    }
    return false
}

func CheckDraw(board [][]string) bool {
    for i := 0; i < rows; i++ {
        for j := 0; j < cols; j++ {
            if board[i][j] == " " {
                return false // If any cell is empty, the game is not a draw yet
            }
        }
    }
    return true // If all cells are occupied and there's no winner, the game is a draw
}

var game = Game{
    Board: InitializeBoard(),
    Turn:  "X",
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
    var move struct {
        Row int `json:"row"`
        Col int `json:"col"`
    }

    err := json.NewDecoder(r.Body).Decode(&move)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    game.Lock.Lock()
    defer game.Lock.Unlock()

    if move.Row == -1 && move.Col == -1 {
        // Initial request to get the initial board state
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(game)
        return
    }

    if move.Row >= 0 && move.Row < rows && move.Col >= 0 && move.Col < cols && game.Board[move.Row][move.Col] == " " && game.Winner == "" {
        MakeMove(game.Board, move.Row, move.Col, game.Turn)
        if CheckWin(game.Board, game.Turn) {
            game.Winner = game.Turn
        } else if CheckDraw(game.Board) {
            game.Draw = true
        } else {
            if game.Turn == "X" {
                game.Turn = "O"
            } else {
                game.Turn = "X"
            }
        }
    } else {
        http.Error(w, "Invalid move", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(game)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
    html, err := ioutil.ReadFile("index.html")
    if err != nil {
        http.Error(w, "Could not read index.html", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    w.Write(html)
}

func main() {
    http.HandleFunc("/", serveHTML)
    http.HandleFunc("/move", moveHandler)
    log.Println("Server is running on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
