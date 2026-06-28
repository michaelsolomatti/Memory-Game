// memory.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Threading;
using System.Diagnostics;

class Memory
{
    static string Colorize(string text, string color)
    {
        string col = color switch
        {
            "red" => "\x1b[91m",
            "green" => "\x1b[92m",
            "yellow" => "\x1b[93m",
            "blue" => "\x1b[94m",
            "magenta" => "\x1b[95m",
            "cyan" => "\x1b[96m",
            "bold" => "\x1b[1m",
            _ => "\x1b[0m"
        };
        return col + text + "\x1b[0m";
    }

    static string[] symbols = {"A","B","C","D","E","F","G","H","I","J","K","L","M",
                               "N","O","P","Q","R","S","T","U","V","W","X","Y","Z",
                               "1","2","3","4","5","6","7","8","9"};
    static string[] symbolColors = {"red", "green", "yellow", "blue", "magenta", "cyan"};

    static string GetSymbolColor(string sym)
    {
        return symbolColors[Math.Abs(sym.GetHashCode()) % symbolColors.Length];
    }

    static string ConfigFile => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".memory_records.json");

    static Dictionary<string, Record> LoadRecords()
    {
        if (!File.Exists(ConfigFile)) return new Dictionary<string, Record>();
        string json = File.ReadAllText(ConfigFile);
        return JsonSerializer.Deserialize<Dictionary<string, Record>>(json) ?? new Dictionary<string, Record>();
    }

    static void SaveRecords(Dictionary<string, Record> records)
    {
        string json = JsonSerializer.Serialize(records, new JsonSerializerOptions { WriteIndented = true });
        File.WriteAllText(ConfigFile, json);
    }

    class Record
    {
        public int Moves { get; set; }
        public int Time { get; set; }
        public string Date { get; set; }
    }

    static void ClearScreen() => Console.Clear();

    static string[,] GenerateBoard(int size)
    {
        int numPairs = (size * size) / 2;
        var selected = symbols[0..numPairs];
        var list = new List<string>();
        foreach (var s in selected) { list.Add(s); list.Add(s); }
        // shuffle
        var rnd = new Random();
        for (int i = list.Count - 1; i > 0; i--)
        {
            int j = rnd.Next(i + 1);
            var tmp = list[i]; list[i] = list[j]; list[j] = tmp;
        }
        var board = new string[size, size];
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                board[i, j] = list[i * size + j];
        return board;
    }

    static void DisplayBoard(string[,] board, bool[,] revealed, int size)
    {
        Console.Write(Colorize("\n  ", "bold"));
        for (int j = 0; j < size; j++) Console.Write(Colorize($"   {j+1}", "bold"));
        Console.WriteLine();
        for (int i = 0; i < size; i++)
        {
            Console.Write(Colorize($"{i+1} ", "bold"));
            for (int j = 0; j < size; j++)
            {
                if (revealed[i, j])
                {
                    string sym = board[i, j];
                    string col = GetSymbolColor(sym);
                    Console.Write(Colorize($" {sym} ", col));
                }
                else Console.Write(Colorize(" ■ ", "bold"));
            }
            Console.WriteLine();
        }
    }

    static (int,int) GetCoordinates(int size)
    {
        while (true)
        {
            Console.Write("Выберите карточку (строка столбец) или q для выхода: ");
            string input = Console.ReadLine().Trim().ToLower();
            if (input == "q") { Console.WriteLine(Colorize("Выход из игры.", "yellow")); Environment.Exit(0); }
            var parts = input.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
            if (parts.Length == 2 && int.TryParse(parts[0], out int row) && int.TryParse(parts[1], out int col))
            {
                if (row >= 1 && row <= size && col >= 1 && col <= size)
                    return (row-1, col-1);
            }
            Console.WriteLine(Colorize("Неверный ввод. Введите два числа через пробел.", "red"));
        }
    }

    static void Main(string[] args)
    {
        int size = 4, timeout = 0;
        for (int i = 0; i < args.Length; i++)
        {
            if (args[i] == "-s" && i+1 < args.Length) size = int.Parse(args[++i]);
            else if (args[i] == "-l" && i+1 < args.Length)
            {
                string level = args[++i];
                if (level == "easy") size = 4;
                else if (level == "medium") size = 6;
                else if (level == "hard") size = 8;
            }
            else if (args[i] == "-t" && i+1 < args.Length) timeout = int.Parse(args[++i]);
            else if (args[i] == "-h")
            {
                Console.WriteLine("Usage: memory [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move");
                return;
            }
        }
        if (size % 2 != 0) { Console.WriteLine(Colorize("Размер должен быть чётным.", "red")); return; }

        var records = LoadRecords();
        string key = size.ToString();

        var board = GenerateBoard(size);
        var revealed = new bool[size, size];
        int moves = 0, pairsFound = 0;
        int totalPairs = (size * size) / 2;
        var stopwatch = Stopwatch.StartNew();

        while (pairsFound < totalPairs)
        {
            ClearScreen();
            Console.WriteLine(Colorize($"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairsFound}/{totalPairs}", "bold"));
            DisplayBoard(board, revealed, size);

            var (r1, c1) = GetCoordinates(size);
            if (revealed[r1, c1]) { Console.WriteLine(Colorize("Карточка уже открыта.", "yellow")); continue; }
            revealed[r1, c1] = true;
            ClearScreen();
            Console.WriteLine(Colorize($"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairsFound}/{totalPairs}", "bold"));
            DisplayBoard(board, revealed, size);

            var (r2, c2) = GetCoordinates(size);
            if ((r2 == r1 && c2 == c1) || revealed[r2, c2])
            {
                Console.WriteLine(Colorize("Неверный выбор.", "yellow"));
                revealed[r1, c1] = false;
                continue;
            }
            revealed[r2, c2] = true;
            moves++;
            ClearScreen();
            Console.WriteLine(Colorize($"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairsFound}/{totalPairs}", "bold"));
            DisplayBoard(board, revealed, size);

            if (board[r1, c1] == board[r2, c2])
            {
                Console.WriteLine(Colorize("✅ Пара найдена!", "green"));
                pairsFound++;
                Thread.Sleep(1000);
            }
            else
            {
                Console.WriteLine(Colorize("❌ Не совпало.", "red"));
                Thread.Sleep(timeout > 0 ? timeout * 1000 : 1500);
                revealed[r1, c1] = false;
                revealed[r2, c2] = false;
            }
        }

        stopwatch.Stop();
        int elapsed = (int)stopwatch.Elapsed.TotalSeconds;
        Console.WriteLine(Colorize($"\n🎉 Поздравляем! Вы завершили игру за {moves} ходов и {elapsed} секунд.", "green"));

        records.TryGetValue(key, out Record best);
        if (best == null || moves < best.Moves || (moves == best.Moves && elapsed < best.Time))
        {
            records[key] = new Record { Moves = moves, Time = elapsed, Date = DateTime.Now.ToString("o") };
            SaveRecords(records);
            Console.WriteLine(Colorize($"🏆 Новый рекорд для размера {size}×{size}!", "yellow"));
        }
        else
        {
            Console.WriteLine(Colorize($"Лучший результат для этого размера: {best.Moves} ходов за {best.Time} сек.", "blue"));
        }
    }
}
