// memory.js
#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');
const { promisify } = require('util');

const COLORS = {
    red: '\x1b[91m',
    green: '\x1b[92m',
    yellow: '\x1b[93m',
    blue: '\x1b[94m',
    magenta: '\x1b[95m',
    cyan: '\x1b[96m',
    reset: '\x1b[0m',
    bold: '\x1b[1m'
};

function colorize(text, color) {
    return COLORS[color] + text + COLORS.reset;
}

const SYMBOLS = ['A','B','C','D','E','F','G','H','I','J','K','L','M',
                 'N','O','P','Q','R','S','T','U','V','W','X','Y','Z',
                 '1','2','3','4','5','6','7','8','9'];
const SYMBOL_COLORS = ['red', 'green', 'yellow', 'blue', 'magenta', 'cyan'];

function getSymbolColor(sym) {
    return SYMBOL_COLORS[hashString(sym) % SYMBOL_COLORS.length];
}

function hashString(str) {
    let h = 0;
    for (let i = 0; i < str.length; i++) {
        h = (h * 31 + str.charCodeAt(i)) & 0xFFFFFFFF;
    }
    return h;
}

const RECORD_FILE = path.join(os.homedir(), '.memory_records.json');

function loadRecords() {
    try {
        return JSON.parse(fs.readFileSync(RECORD_FILE, 'utf8'));
    } catch { return {}; }
}

function saveRecords(records) {
    fs.writeFileSync(RECORD_FILE, JSON.stringify(records, null, 2));
}

function clearScreen() {
    console.clear();
}

function generateBoard(size) {
    const numPairs = (size * size) / 2;
    const selected = SYMBOLS.slice(0, numPairs);
    let board = [];
    for (const s of selected) {
        board.push(s, s);
    }
    // shuffle
    for (let i = board.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [board[i], board[j]] = [board[j], board[i]];
    }
    const result = [];
    for (let i = 0; i < size; i++) {
        result[i] = [];
        for (let j = 0; j < size; j++) {
            result[i][j] = board[i * size + j];
        }
    }
    return result;
}

function displayBoard(board, revealed, size) {
    console.log(colorize('\n  ', 'bold') + '   ' + Array.from({length: size}, (_,i) => colorize(i+1, 'bold')).join('   '));
    for (let i = 0; i < size; i++) {
        let row = colorize(i+1, 'bold') + ' ';
        for (let j = 0; j < size; j++) {
            if (revealed[i][j]) {
                const sym = board[i][j];
                const col = getSymbolColor(sym);
                row += colorize(` ${sym} `, col);
            } else {
                row += colorize(' ■ ', 'bold');
            }
        }
        console.log(row);
    }
}

async function getCoordinates(size, rl) {
    return new Promise((resolve) => {
        const prompt = () => {
            rl.question('Выберите карточку (строка столбец) или q для выхода: ', (input) => {
                input = input.trim().toLowerCase();
                if (input === 'q') {
                    console.log(colorize('Выход из игры.', 'yellow'));
                    process.exit(0);
                }
                const parts = input.split(/\s+/);
                if (parts.length === 2) {
                    const row = parseInt(parts[0]);
                    const col = parseInt(parts[1]);
                    if (!isNaN(row) && !isNaN(col) && row >= 1 && row <= size && col >= 1 && col <= size) {
                        resolve([row-1, col-1]);
                        return;
                    }
                }
                console.log(colorize('Неверный ввод. Введите два числа через пробел.', 'red'));
                prompt();
            });
        };
        prompt();
    });
}

async function main() {
    const args = process.argv.slice(2);
    let size = 4;
    let timeout = 0;
    for (let i = 0; i < args.length; i++) {
        if (args[i] === '-s' && i+1 < args.length) {
            size = parseInt(args[++i]);
        } else if (args[i] === '-l' && i+1 < args.length) {
            const level = args[++i];
            if (level === 'easy') size = 4;
            else if (level === 'medium') size = 6;
            else if (level === 'hard') size = 8;
        } else if (args[i] === '-t' && i+1 < args.length) {
            timeout = parseInt(args[++i]);
        } else if (args[i] === '-h') {
            console.log('Usage: node memory.js [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move');
            return;
        }
    }
    if (size % 2 !== 0) {
        console.log(colorize('Размер должен быть чётным.', 'red'));
        return;
    }

    const records = loadRecords();
    const key = String(size);

    const board = generateBoard(size);
    const revealed = Array.from({length: size}, () => Array(size).fill(false));
    let moves = 0;
    let pairsFound = 0;
    const totalPairs = (size * size) / 2;
    const start = Date.now();

    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    });

    while (pairsFound < totalPairs) {
        clearScreen();
        console.log(colorize(`🧠  МЕМОРИ  |  Размер ${size}×${size}  |  Ходы: ${moves}  |  Пары: ${pairsFound}/${totalPairs}`, 'bold'));
        displayBoard(board, revealed, size);

        // Первая карточка
        const [r1, c1] = await getCoordinates(size, rl);
        if (revealed[r1][c1]) {
            console.log(colorize('Карточка уже открыта.', 'yellow'));
            continue;
        }
        revealed[r1][c1] = true;
        clearScreen();
        console.log(colorize(`🧠  МЕМОРИ  |  Размер ${size}×${size}  |  Ходы: ${moves}  |  Пары: ${pairsFound}/${totalPairs}`, 'bold'));
        displayBoard(board, revealed, size);

        // Вторая карточка
        const [r2, c2] = await getCoordinates(size, rl);
        if ((r2 === r1 && c2 === c1) || revealed[r2][c2]) {
            console.log(colorize('Неверный выбор.', 'yellow'));
            revealed[r1][c1] = false;
            continue;
        }
        revealed[r2][c2] = true;
        moves++;
        clearScreen();
        console.log(colorize(`🧠  МЕМОРИ  |  Размер ${size}×${size}  |  Ходы: ${moves}  |  Пары: ${pairsFound}/${totalPairs}`, 'bold'));
        displayBoard(board, revealed, size);

        if (board[r1][c1] === board[r2][c2]) {
            console.log(colorize('✅ Пара найдена!', 'green'));
            pairsFound++;
            await new Promise(r => setTimeout(r, 1000));
        } else {
            console.log(colorize('❌ Не совпало.', 'red'));
            await new Promise(r => setTimeout(r, timeout > 0 ? timeout * 1000 : 1500));
            revealed[r1][c1] = false;
            revealed[r2][c2] = false;
        }
    }

    const elapsed = Math.floor((Date.now() - start) / 1000);
    console.log(colorize(`\n🎉 Поздравляем! Вы завершили игру за ${moves} ходов и ${elapsed} секунд.`, 'green'));

    const best = records[key];
    if (!best || moves < best.moves || (moves === best.moves && elapsed < best.time)) {
        records[key] = { moves, time: elapsed, date: new Date().toISOString() };
        saveRecords(records);
        console.log(colorize(`🏆 Новый рекорд для размера ${size}×${size}!`, 'yellow'));
    } else {
        console.log(colorize(`Лучший результат для этого размера: ${best.moves} ходов за ${best.time} сек.`, 'blue'));
    }
    rl.close();
}

main().catch(err => console.error(err));
