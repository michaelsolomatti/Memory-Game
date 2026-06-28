// memory.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	reset   = "\033[0m"
	red     = "\033[91m"
	green   = "\033[92m"
	yellow  = "\033[93m"
	blue    = "\033[94m"
	magenta = "\033[95m"
	cyan    = "\033[96m"
	bold    = "\033[1m"
)

func colorize(text, color string) string {
	return color + text + reset
}

var symbols = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
	"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"1", "2", "3", "4", "5", "6", "7", "8", "9"}
var symbolColors = []string{red, green, yellow, blue, magenta, cyan}

func getSymbolColor(sym string) string {
	return symbolColors[hash(sym)%len(symbolColors)]
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	return h
}

type Record struct {
	Moves int    `json:"moves"`
	Time  int    `json:"time"`
	Date  string `json:"date"`
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func getRecordFile() string {
	return getHomeDir() + "/.memory_records.json"
}

func loadRecords() map[string]Record {
	data, err := os.ReadFile(getRecordFile())
	if err != nil {
		return make(map[string]Record)
	}
	var records map[string]Record
	json.Unmarshal(data, &records)
	return records
}

func saveRecords(records map[string]Record) {
	data, _ := json.MarshalIndent(records, "", "  ")
	os.WriteFile(getRecordFile(), data, 0644)
}

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func generateBoard(size int) [][]string {
	numPairs := (size * size) / 2
	selected := symbols[:numPairs]
	var board []string
	for _, s := range selected {
		board = append(board, s, s)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(board), func(i, j int) { board[i], board[j] = board[j], board[i] })
	result := make([][]string, size)
	for i := range result {
		result[i] = make([]string, size)
		for j := range result[i] {
			result[i][j] = board[i*size+j]
		}
	}
	return result
}

func displayBoard(board [][]string, revealed [][]bool, size int) {
	fmt.Print(colorize("\n  ", bold))
	for j := 0; j < size; j++ {
		fmt.Print(colorize(fmt.Sprintf("   %d", j+1), bold))
	}
	fmt.Println()
	for i := 0; i < size; i++ {
		fmt.Print(colorize(fmt.Sprintf("%d ", i+1), bold))
		for j := 0; j < size; j++ {
			if revealed[i][j] {
				sym := board[i][j]
				col := getSymbolColor(sym)
				fmt.Print(colorize(" "+sym+" ", col))
			} else {
				fmt.Print(colorize(" ■ ", bold))
			}
		}
		fmt.Println()
	}
}

func getCoordinates(size int) (int, int) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Выберите карточку (строка столбец) или 'q' для выхода: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" || input == "Q" {
			fmt.Println(colorize("Выход из игры.", yellow))
			os.Exit(0)
		}
		parts := strings.Fields(input)
		if len(parts) == 2 {
			row, err1 := strconv.Atoi(parts[0])
			col, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && row >= 1 && row <= size && col >= 1 && col <= size {
				return row - 1, col - 1
			}
		}
		fmt.Println(colorize("Неверный ввод. Введите два числа через пробел.", red))
	}
}

func main() {
	size := 4
	timeout := 0
	level := ""

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-s":
			if i+1 < len(os.Args) {
				size, _ = strconv.Atoi(os.Args[i+1])
				i++
			}
		case "-l":
			if i+1 < len(os.Args) {
				level = os.Args[i+1]
				switch level {
				case "easy":
					size = 4
				case "medium":
					size = 6
				case "hard":
					size = 8
				}
				i++
			}
		case "-t":
			if i+1 < len(os.Args) {
				timeout, _ = strconv.Atoi(os.Args[i+1])
				i++
			}
		case "-h":
			fmt.Println("Usage: memory [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move")
			return
		}
	}
	if size%2 != 0 {
		fmt.Println(colorize("Размер должен быть чётным.", red))
		return
	}

	records := loadRecords()
	key := strconv.Itoa(size)

	board := generateBoard(size)
	revealed := make([][]bool, size)
	for i := range revealed {
		revealed[i] = make([]bool, size)
	}
	moves := 0
	pairsFound := 0
	totalPairs := (size * size) / 2
	start := time.Now()

	for pairsFound < totalPairs {
		clearScreen()
		fmt.Println(colorize(fmt.Sprintf("🧠  МЕМОРИ  |  Размер %d×%d  |  Ходы: %d  |  Пары: %d/%d",
			size, size, moves, pairsFound, totalPairs), bold))
		displayBoard(board, revealed, size)

		// Первая карточка
		r1, c1 := getCoordinates(size)
		if revealed[r1][c1] {
			fmt.Println(colorize("Карточка уже открыта.", yellow))
			continue
		}
		revealed[r1][c1] = true
		clearScreen()
		fmt.Println(colorize(fmt.Sprintf("🧠  МЕМОРИ  |  Размер %d×%d  |  Ходы: %d  |  Пары: %d/%d",
			size, size, moves, pairsFound, totalPairs), bold))
		displayBoard(board, revealed, size)

		// Вторая карточка
		r2, c2 := getCoordinates(size)
		if (r2 == r1 && c2 == c1) || revealed[r2][c2] {
			fmt.Println(colorize("Неверный выбор.", yellow))
			revealed[r1][c1] = false
			continue
		}
		revealed[r2][c2] = true
		moves++
		clearScreen()
		fmt.Println(colorize(fmt.Sprintf("🧠  МЕМОРИ  |  Размер %d×%d  |  Ходы: %d  |  Пары: %d/%d",
			size, size, moves, pairsFound, totalPairs), bold))
		displayBoard(board, revealed, size)

		if board[r1][c1] == board[r2][c2] {
			fmt.Println(colorize("✅ Пара найдена!", green))
			pairsFound++
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println(colorize("❌ Не совпало.", red))
			time.Sleep(time.Duration(timeout) * time.Second)
			if timeout == 0 {
				time.Sleep(1 * time.Second)
			}
			revealed[r1][c1] = false
			revealed[r2][c2] = false
		}
	}

	elapsed := int(time.Since(start).Seconds())
	fmt.Println(colorize(fmt.Sprintf("\n🎉 Поздравляем! Вы завершили игру за %d ходов и %d секунд.", moves, elapsed), green))

	best, ok := records[key]
	if !ok || moves < best.Moves || (moves == best.Moves && elapsed < best.Time) {
		records[key] = Record{Moves: moves, Time: elapsed, Date: time.Now().Format(time.RFC3339)}
		saveRecords(records)
		fmt.Println(colorize(fmt.Sprintf("🏆 Новый рекорд для размера %d×%d!", size, size), yellow))
	} else {
		fmt.Println(colorize(fmt.Sprintf("Лучший результат для этого размера: %d ходов за %d сек.", best.Moves, best.Time), blue))
	}
}
