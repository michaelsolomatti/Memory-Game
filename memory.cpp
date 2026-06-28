// memory.cpp
#include <iostream>
#include <vector>
#include <string>
#include <algorithm>
#include <random>
#include <chrono>
#include <thread>
#include <fstream>
#include <json/json.h> // sudo apt-get install libjsoncpp-dev

using namespace std;

const string RESET = "\033[0m";
const string RED = "\033[91m";
const string GREEN = "\033[92m";
const string YELLOW = "\033[93m";
const string BLUE = "\033[94m";
const string MAGENTA = "\033[95m";
const string CYAN = "\033[96m";
const string BOLD = "\033[1m";

string colorize(const string& text, const string& color) {
    return color + text + RESET;
}

vector<string> symbols = {"A","B","C","D","E","F","G","H","I","J","K","L","M",
                          "N","O","P","Q","R","S","T","U","V","W","X","Y","Z",
                          "1","2","3","4","5","6","7","8","9"};
vector<string> symbolColors = {RED, GREEN, YELLOW, BLUE, MAGENTA, CYAN};

string getSymbolColor(const string& sym) {
    return symbolColors[hash<string>{}(sym) % symbolColors.size()];
}

string getHomeDir() {
    const char* home = getenv("HOME");
    if (!home) home = getenv("USERPROFILE");
    return string(home);
}

string getRecordFile() {
    return getHomeDir() + "/.memory_records.json";
}

Json::Value loadRecords() {
    ifstream f(getRecordFile());
    Json::Value root;
    if (!f) return root;
    f >> root;
    return root;
}

void saveRecords(const Json::Value& records) {
    ofstream f(getRecordFile());
    f << records.toStyledString();
}

vector<vector<string>> generateBoard(int size) {
    int numPairs = (size * size) / 2;
    vector<string> selected(symbols.begin(), symbols.begin() + numPairs);
    vector<string> board;
    for (string s : selected) {
        board.push_back(s);
        board.push_back(s);
    }
    random_device rd;
    mt19937 g(rd());
    shuffle(board.begin(), board.end(), g);
    vector<vector<string>> result(size, vector<string>(size));
    for (int i = 0; i < size; ++i)
        for (int j = 0; j < size; ++j)
            result[i][j] = board[i * size + j];
    return result;
}

void displayBoard(const vector<vector<string>>& board, const vector<vector<bool>>& revealed, int size) {
    cout << colorize("\n  ", BOLD);
    for (int j = 0; j < size; ++j) cout << colorize("   " + to_string(j+1), BOLD);
    cout << endl;
    for (int i = 0; i < size; ++i) {
        cout << colorize(to_string(i+1) + " ", BOLD);
        for (int j = 0; j < size; ++j) {
            if (revealed[i][j]) {
                string sym = board[i][j];
                string col = getSymbolColor(sym);
                cout << colorize(" " + sym + " ", col);
            } else {
                cout << colorize(" ■ ", BOLD);
            }
        }
        cout << endl;
    }
}

void clearScreen() {
    cout << "\033[2J\033[1;1H";
}

pair<int,int> getCoordinates(int size) {
    while (true) {
        string input;
        cout << "Выберите карточку (строка столбец) или 'q' для выхода: ";
        cin >> input;
        if (input == "q" || input == "Q") {
            cout << colorize("Выход из игры.", YELLOW) << endl;
            exit(0);
        }
        int row, col;
        if (sscanf(input.c_str(), "%d %d", &row, &col) == 2) {
            if (row >= 1 && row <= size && col >= 1 && col <= size) {
                return {row-1, col-1};
            }
        }
        cout << colorize("Неверный ввод. Введите два числа через пробел.", RED) << endl;
    }
}

int main(int argc, char* argv[]) {
    int size = 4;
    int timeout = 0;
    string level;

    for (int i = 1; i < argc; ++i) {
        string arg = argv[i];
        if (arg == "-s" && i+1 < argc) size = stoi(argv[++i]);
        else if (arg == "-l" && i+1 < argc) {
            level = argv[++i];
            if (level == "easy") size = 4;
            else if (level == "medium") size = 6;
            else if (level == "hard") size = 8;
        }
        else if (arg == "-t" && i+1 < argc) timeout = stoi(argv[++i]);
        else if (arg == "-h") {
            cout << "Usage: memory [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move\n";
            return 0;
        }
    }
    if (size % 2 != 0) {
        cout << colorize("Размер должен быть чётным.", RED) << endl;
        return 1;
    }

    Json::Value records = loadRecords();
    string key = to_string(size);

    auto board = generateBoard(size);
    vector<vector<bool>> revealed(size, vector<bool>(size, false));
    int moves = 0;
    int pairsFound = 0;
    int totalPairs = (size * size) / 2;
    auto start = chrono::steady_clock::now();

    while (pairsFound < totalPairs) {
        clearScreen();
        cout << colorize("🧠  МЕМОРИ  |  Размер " + to_string(size) + "×" + to_string(size) +
                         "  |  Ходы: " + to_string(moves) + "  |  Пары: " + to_string(pairsFound) + "/" + to_string(totalPairs), BOLD) << endl;
        displayBoard(board, revealed, size);

        // Первая карточка
        auto [r1, c1] = getCoordinates(size);
        if (revealed[r1][c1]) {
            cout << colorize("Карточка уже открыта.", YELLOW) << endl;
            continue;
        }
        revealed[r1][c1] = true;
        clearScreen();
        cout << colorize("🧠  МЕМОРИ  |  Размер " + to_string(size) + "×" + to_string(size) +
                         "  |  Ходы: " + to_string(moves) + "  |  Пары: " + to_string(pairsFound) + "/" + to_string(totalPairs), BOLD) << endl;
        displayBoard(board, revealed, size);

        // Вторая карточка
        auto [r2, c2] = getCoordinates(size);
        if ((r2 == r1 && c2 == c1) || revealed[r2][c2]) {
            cout << colorize("Неверный выбор.", YELLOW) << endl;
            revealed[r1][c1] = false;
            continue;
        }
        revealed[r2][c2] = true;
        moves++;
        clearScreen();
        cout << colorize("🧠  МЕМОРИ  |  Размер " + to_string(size) + "×" + to_string(size) +
                         "  |  Ходы: " + to_string(moves) + "  |  Пары: " + to_string(pairsFound) + "/" + to_string(totalPairs), BOLD) << endl;
        displayBoard(board, revealed, size);

        if (board[r1][c1] == board[r2][c2]) {
            cout << colorize("✅ Пара найдена!", GREEN) << endl;
            pairsFound++;
            this_thread::sleep_for(chrono::seconds(1));
        } else {
            cout << colorize("❌ Не совпало.", RED) << endl;
            this_thread::sleep_for(chrono::seconds(timeout > 0 ? timeout : 1));
            revealed[r1][c1] = false;
            revealed[r2][c2] = false;
        }
    }

    auto end = chrono::steady_clock::now();
    int elapsed = chrono::duration_cast<chrono::seconds>(end - start).count();
    cout << colorize("\n🎉 Поздравляем! Вы завершили игру за " + to_string(moves) + " ходов и " + to_string(elapsed) + " секунд.", GREEN) << endl;

    // Обновление рекорда
    Json::Value best = records[key];
    if (best.isNull() || moves < best["moves"].asInt() || (moves == best["moves"].asInt() && elapsed < best["time"].asInt())) {
        best["moves"] = moves;
        best["time"] = elapsed;
        time_t now = time(nullptr);
        char buf[64];
        strftime(buf, sizeof(buf), "%Y-%m-%dT%H:%M:%S", localtime(&now));
        best["date"] = string(buf);
        records[key] = best;
        saveRecords(records);
        cout << colorize("🏆 Новый рекорд для размера " + to_string(size) + "×" + to_string(size) + "!", YELLOW) << endl;
    } else {
        cout << colorize("Лучший результат для этого размера: " + to_string(best["moves"].asInt()) + " ходов за " + to_string(best["time"].asInt()) + " сек.", BLUE) << endl;
    }
    return 0;
}
