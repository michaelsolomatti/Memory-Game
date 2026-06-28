# memory.py
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import os
import random
import json
import time
import argparse
from datetime import datetime
from pathlib import Path

# ANSI colors
COLORS = {
    'red': '\033[91m',
    'green': '\033[92m',
    'yellow': '\033[93m',
    'blue': '\033[94m',
    'magenta': '\033[95m',
    'cyan': '\033[96m',
    'reset': '\033[0m',
    'bold': '\033[1m'
}

def colorize(text, color):
    return f"{COLORS.get(color, '')}{text}{COLORS['reset']}"

# Символы для карточек (разные наборы)
SYMBOLS = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
           'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
           '1', '2', '3', '4', '5', '6', '7', '8', '9']

# Цвета для символов
SYMBOL_COLORS = ['red', 'green', 'yellow', 'blue', 'magenta', 'cyan']

def get_symbol_color(symbol):
    return SYMBOL_COLORS[hash(symbol) % len(SYMBOL_COLORS)]

RECORD_FILE = Path.home() / '.memory_records.json'

def load_records():
    if RECORD_FILE.exists():
        with open(RECORD_FILE, 'r') as f:
            return json.load(f)
    return {}

def save_records(records):
    with open(RECORD_FILE, 'w') as f:
        json.dump(records, f, indent=2)

def generate_board(size):
    """Генерирует поле с парами символов и перемешивает."""
    num_pairs = (size * size) // 2
    symbols = random.sample(SYMBOLS, num_pairs)
    board = symbols * 2
    random.shuffle(board)
    return [board[i*size:(i+1)*size] for i in range(size)]

def display_board(board, revealed, size):
    """Выводит игровое поле."""
    print(colorize("\n  " + "   ".join(str(i+1) for i in range(size)), 'bold'))
    for i in range(size):
        row_display = [str(i+1) + " "]
        for j in range(size):
            if revealed[i][j]:
                symbol = board[i][j]
                color = get_symbol_color(symbol)
                row_display.append(colorize(f" {symbol} ", color))
            else:
                row_display.append(colorize(" ■ ", 'bold'))
        print(" ".join(row_display))

def get_coordinates(size, prompt):
    while True:
        try:
            inp = input(prompt).strip()
            if inp.lower() == 'q':
                print(colorize("Выход из игры.", 'yellow'))
                sys.exit(0)
            row, col = map(int, inp.split())
            if 1 <= row <= size and 1 <= col <= size:
                return row-1, col-1
            print(colorize("Координаты вне поля. Попробуйте снова.", 'red'))
        except ValueError:
            print(colorize("Введите два числа через пробел (строка столбец).", 'red'))

def main():
    parser = argparse.ArgumentParser(description="Memory Game – найди пару")
    parser.add_argument('-s', '--size', type=int, default=4, help='Размер поля N×N')
    parser.add_argument('-l', '--level', choices=['easy', 'medium', 'hard'],
                        help='Уровень сложности (easy=4, medium=6, hard=8)')
    parser.add_argument('-t', '--timeout', type=int, default=0, help='Таймаут на ход (сек)')
    args = parser.parse_args()

    # Определение размера
    size = args.size
    if args.level:
        size = {'easy': 4, 'medium': 6, 'hard': 8}[args.level]

    if size % 2 != 0:
        print(colorize("Размер должен быть чётным.", 'red'))
        sys.exit(1)

    records = load_records()
    key = str(size)

    board = generate_board(size)
    revealed = [[False] * size for _ in range(size)]
    moves = 0
    start_time = time.time()
    pairs_found = 0
    total_pairs = (size * size) // 2

    while pairs_found < total_pairs:
        os.system('clear' if os.name == 'posix' else 'cls')
        print(colorize(f"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairs_found}/{total_pairs}", 'bold'))
        display_board(board, revealed, size)

        # Выбор первой карточки
        row1, col1 = get_coordinates(size, "Выберите первую карточку (строка столбец): ")
        if revealed[row1][col1]:
            print(colorize("Эта карточка уже открыта.", 'yellow'))
            continue
        revealed[row1][col1] = True
        os.system('clear' if os.name == 'posix' else 'cls')
        print(colorize(f"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairs_found}/{total_pairs}", 'bold'))
        display_board(board, revealed, size)

        # Выбор второй карточки
        row2, col2 = get_coordinates(size, "Выберите вторую карточку (строка столбец): ")
        if (row2, col2) == (row1, col1) or revealed[row2][col2]:
            print(colorize("Неверный выбор или карточка уже открыта.", 'yellow'))
            revealed[row1][col1] = False
            continue

        revealed[row2][col2] = True
        moves += 1
        os.system('clear' if os.name == 'posix' else 'cls')
        print(colorize(f"🧠  МЕМОРИ  |  Размер {size}×{size}  |  Ходы: {moves}  |  Пары: {pairs_found}/{total_pairs}", 'bold'))
        display_board(board, revealed, size)

        # Проверка пары
        if board[row1][col1] == board[row2][col2]:
            print(colorize("✅ Пара найдена!", 'green'))
            pairs_found += 1
            # Если таймаут не нужен, просто задержка для визуализации
            time.sleep(1)
        else:
            print(colorize("❌ Не совпало.", 'red'))
            time.sleep(1.5 if args.timeout == 0 else args.timeout)
            revealed[row1][col1] = False
            revealed[row2][col2] = False

    elapsed = int(time.time() - start_time)
    print(colorize(f"\n🎉 Поздравляем! Вы завершили игру за {moves} ходов и {elapsed} секунд.", 'green'))

    # Обновление рекорда
    best = records.get(key, {})
    if not best or moves < best.get('moves', float('inf')) or (moves == best.get('moves') and elapsed < best.get('time', float('inf'))):
        best = {'moves': moves, 'time': elapsed, 'date': datetime.now().isoformat()}
        records[key] = best
        save_records(records)
        print(colorize(f"🏆 Новый рекорд для размера {size}×{size}!", 'yellow'))
    else:
        print(colorize(f"Лучший результат для этого размера: {best['moves']} ходов за {best['time']} сек.", 'blue'))

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print(colorize("\n👋 Игра прервана.", 'yellow'))
        sys.exit(0)
