// memory.java
import java.io.*;
import java.nio.file.*;
import java.util.*;
import java.util.concurrent.*;
import com.google.gson.*;

public class memory {
    private static final String RESET = "\u001B[0m";
    private static final String RED = "\u001B[91m";
    private static final String GREEN = "\u001B[92m";
    private static final String YELLOW = "\u001B[93m";
    private static final String BLUE = "\u001B[94m";
    private static final String MAGENTA = "\u001B[95m";
    private static final String CYAN = "\u001B[96m";
    private static final String BOLD = "\u001B[1m";

    private static String colorize(String text, String color) {
        return color + text + RESET;
    }

    private static String[] symbols = {"A","B","C","D","E","F","G","H","I","J","K","L","M",
                                       "N","O","P","Q","R","S","T","U","V","W","X","Y","Z",
                                       "1","2","3","4","5","6","7","8","9"};
    private static String[] symbolColors = {RED, GREEN, YELLOW, BLUE, MAGENTA, CYAN};

    private static String getSymbolColor(String sym) {
        return symbolColors[Math.abs(sym.hashCode()) % symbolColors.length];
    }

    private static String configFile = System.getProperty("user.home") + "/.memory_records.json";

    private static Map<String, Record> loadRecords() throws IOException {
        Path path = Paths.get(configFile);
        if (!Files.exists(path)) return new HashMap<>();
        String json = new String(Files.readAllBytes(path));
        Gson gson = new Gson();
        Type type = new com.google.gson.reflect.TypeToken<Map<String, Record>>(){}.getType();
        return gson.fromJson(json, type);
    }

    private static void saveRecords(Map<String, Record> records) throws IOException {
        Gson gson = new GsonBuilder().setPrettyPrinting().create();
        String json = gson.toJson(records);
        Files.write(Paths.get(configFile), json.getBytes());
    }

    static class Record {
        int moves;
        int time;
        String date;
    }

    private static void clearScreen() {
        System.out.print("\033[H\033[2J");
        System.out.flush();
    }

    private static String[][] generateBoard(int size) {
        int numPairs = (size * size) / 2;
        List<String> selected = new ArrayList<>();
        for (int i = 0; i < numPairs; i++) selected.add(symbols[i]);
        List<String> boardList = new ArrayList<>();
        for (String s : selected) { boardList.add(s); boardList.add(s); }
        Collections.shuffle(boardList);
        String[][] board = new String[size][size];
        for (int i = 0; i < size; i++)
            for (int j = 0; j < size; j++)
                board[i][j] = boardList.get(i * size + j);
        return board;
    }

    private static void displayBoard(String[][] board, boolean[][] revealed, int size) {
        System.out.print(colorize("\n  ", BOLD));
        for (int j = 0; j < size; j++) System.out.print(colorize("   " + (j+1), BOLD));
        System.out.println();
        for (int i = 0; i < size; i++) {
            System.out.print(colorize((i+1) + " ", BOLD));
            for (int j = 0; j < size; j++) {
                if (revealed[i][j]) {
                    String sym = board[i][j];
                    String col = getSymbolColor(sym);
                    System.out.print(colorize(" " + sym + " ", col));
                } else {
                    System.out.print(colorize(" ■ ", BOLD));
                }
            }
            System.out.println();
        }
    }

    private static int[] getCoordinates(int size) throws IOException {
        BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
        while (true) {
            System.out.print("Выберите карточку (строка столбец) или q для выхода: ");
            String input = reader.readLine().trim().toLowerCase();
            if (input.equals("q")) {
                System.out.println(colorize("Выход из игры.", YELLOW));
                System.exit(0);
            }
            String[] parts = input.split("\\s+");
            if (parts.length == 2) {
                try {
                    int row = Integer.parseInt(parts[0]);
                    int col = Integer.parseInt(parts[1]);
                    if (row >= 1 && row <= size && col >= 1 && col <= size) {
                        return new int[]{row-1, col-1};
                    }
                } catch (NumberFormatException ignored) {}
            }
            System.out.println(colorize("Неверный ввод. Введите два числа через пробел.", RED));
        }
    }

    public static void main(String[] args) throws IOException, InterruptedException {
        int size = 4, timeout = 0;
        for (int i = 0; i < args.length; i++) {
            if (args[i].equals("-s") && i+1 < args.length) size = Integer.parseInt(args[++i]);
            else if (args[i].equals("-l") && i+1 < args.length) {
                String level = args[++i];
                if (level.equals("easy")) size = 4;
                else if (level.equals("medium")) size = 6;
                else if (level.equals("hard")) size = 8;
            } else if (args[i].equals("-t") && i+1 < args.length) timeout = Integer.parseInt(args[++i]);
            else if (args[i].equals("-h")) {
                System.out.println("Usage: java memory [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move");
                return;
            }
        }
        if (size % 2 != 0) {
            System.out.println(colorize("Размер должен быть чётным.", RED));
            return;
        }

        Map<String, Record> records = loadRecords();
        String key = String.valueOf(size);

        String[][] board = generateBoard(size);
        boolean[][] revealed = new boolean[size][size];
        int moves = 0, pairsFound = 0;
        int totalPairs = (size * size) / 2;
        long start = System.currentTimeMillis();

        while (pairsFound < totalPairs) {
            clearScreen();
            System.out.println(colorize("🧠  МЕМОРИ  |  Размер " + size + "×" + size +
                                        "  |  Ходы: " + moves + "  |  Пары: " + pairsFound + "/" + totalPairs, BOLD));
            displayBoard(board, revealed, size);

            int[] pos1 = getCoordinates(size);
            int r1 = pos1[0], c1 = pos1[1];
            if (revealed[r1][c1]) {
                System.out.println(colorize("Карточка уже открыта.", YELLOW));
                continue;
            }
            revealed[r1][c1] = true;
            clearScreen();
            System.out.println(colorize("🧠  МЕМОРИ  |  Размер " + size + "×" + size +
                                        "  |  Ходы: " + moves + "  |  Пары: " + pairsFound + "/" + totalPairs, BOLD));
            displayBoard(board, revealed, size);

            int[] pos2 = getCoordinates(size);
            int r2 = pos2[0], c2 = pos2[1];
            if ((r2 == r1 && c2 == c1) || revealed[r2][c2]) {
                System.out.println(colorize("Неверный выбор.", YELLOW));
                revealed[r1][c1] = false;
                continue;
            }
            revealed[r2][c2] = true;
            moves++;
            clearScreen();
            System.out.println(colorize("🧠  МЕМОРИ  |  Размер " + size + "×" + size +
                                        "  |  Ходы: " + moves + "  |  Пары: " + pairsFound + "/" + totalPairs, BOLD));
            displayBoard(board, revealed, size);

            if (board[r1][c1].equals(board[r2][c2])) {
                System.out.println(colorize("✅ Пара найдена!", GREEN));
                pairsFound++;
                Thread.sleep(1000);
            } else {
                System.out.println(colorize("❌ Не совпало.", RED));
                Thread.sleep(timeout > 0 ? timeout * 1000 : 1500);
                revealed[r1][c1] = false;
                revealed[r2][c2] = false;
            }
        }

        long elapsed = (System.currentTimeMillis() - start) / 1000;
        System.out.println(colorize("\n🎉 Поздравляем! Вы завершили игру за " + moves + " ходов и " + elapsed + " секунд.", GREEN));

        Record best = records.get(key);
        if (best == null || moves < best.moves || (moves == best.moves && elapsed < best.time)) {
            best = new Record();
            best.moves = moves;
            best.time = (int) elapsed;
            best.date = new java.util.Date().toString();
            records.put(key, best);
            saveRecords(records);
            System.out.println(colorize("🏆 Новый рекорд для размера " + size + "×" + size + "!", YELLOW));
        } else {
            System.out.println(colorize("Лучший результат для этого размера: " + best.moves + " ходов за " + best.time + " сек.", BLUE));
        }
    }
}
