#!/usr/bin/env ruby
# memory.rb
# encoding: UTF-8

require 'json'
require 'fileutils'
require 'io/console'

COLORS = {
  red: "\e[91m",
  green: "\e[92m",
  yellow: "\e[93m",
  blue: "\e[94m",
  magenta: "\e[95m",
  cyan: "\e[96m",
  bold: "\e[1m",
  reset: "\e[0m"
}

def colorize(text, color)
  "#{COLORS[color]}#{text}#{COLORS[:reset]}"
end

SYMBOLS = ['A','B','C','D','E','F','G','H','I','J','K','L','M',
           'N','O','P','Q','R','S','T','U','V','W','X','Y','Z',
           '1','2','3','4','5','6','7','8','9']
SYMBOL_COLORS = [:red, :green, :yellow, :blue, :magenta, :cyan]

def get_symbol_color(sym)
  SYMBOL_COLORS[sym.hash.abs % SYMBOL_COLORS.length]
end

RECORD_FILE = File.join(Dir.home, '.memory_records.json')

def load_records
  return {} unless File.exist?(RECORD_FILE)
  JSON.parse(File.read(RECORD_FILE))
end

def save_records(records)
  File.write(RECORD_FILE, JSON.pretty_generate(records))
end

def clear_screen
  system('clear') || system('cls')
end

def generate_board(size)
  num_pairs = (size * size) / 2
  selected = SYMBOLS[0...num_pairs]
  board = (selected + selected).shuffle
  result = []
  (0...size).each do |i|
    row = []
    (0...size).each { |j| row << board[i * size + j] }
    result << row
  end
  result
end

def display_board(board, revealed, size)
  print colorize("\n  ", :bold)
  (1..size).each { |j| print colorize("   #{j}", :bold) }
  puts
  (0...size).each do |i|
    print colorize("#{i+1} ", :bold)
    (0...size).each do |j|
      if revealed[i][j]
        sym = board[i][j]
        col = get_symbol_color(sym)
        print colorize(" #{sym} ", col)
      else
        print colorize(" ■ ", :bold)
      end
    end
    puts
  end
end

def get_coordinates(size)
  loop do
    print "Выберите карточку (строка столбец) или q для выхода: "
    input = STDIN.gets.chomp.strip
    if input.downcase == 'q'
      puts colorize("Выход из игры.", :yellow)
      exit 0
    end
    parts = input.split
    if parts.size == 2 && parts.all? { |p| p =~ /^\d+$/ }
      row = parts[0].to_i
      col = parts[1].to_i
      if row >= 1 && row <= size && col >= 1 && col <= size
        return [row-1, col-1]
      end
    end
    puts colorize("Неверный ввод. Введите два числа через пробел.", :red)
  end
end

def main
  size = 4
  timeout = 0
  i = 0
  while i < ARGV.length
    case ARGV[i]
    when '-s'
      size = ARGV[i+1].to_i
      i += 2
    when '-l'
      level = ARGV[i+1]
      size = case level
             when 'easy' then 4
             when 'medium' then 6
             when 'hard' then 8
             else size
             end
      i += 2
    when '-t'
      timeout = ARGV[i+1].to_i
      i += 2
    when '-h'
      puts "Usage: ruby memory.rb [options]\n  -s <N>    Size NxN\n  -l <level> easy|medium|hard\n  -t <sec>  Timeout per move"
      return
    else
      i += 1
    end
  end

  if size.even? == false
    puts colorize("Размер должен быть чётным.", :red)
    return
  end

  records = load_records
  key = size.to_s

  board = generate_board(size)
  revealed = Array.new(size) { Array.new(size, false) }
  moves = 0
  pairs_found = 0
  total_pairs = (size * size) / 2
  start_time = Time.now

  while pairs_found < total_pairs
    clear_screen
    puts colorize("🧠  МЕМОРИ  |  Размер #{size}×#{size}  |  Ходы: #{moves}  |  Пары: #{pairs_found}/#{total_pairs}", :bold)
    display_board(board, revealed, size)

    r1, c1 = get_coordinates(size)
    if revealed[r1][c1]
      puts colorize("Карточка уже открыта.", :yellow)
      next
    end
    revealed[r1][c1] = true
    clear_screen
    puts colorize("🧠  МЕМОРИ  |  Размер #{size}×#{size}  |  Ходы: #{moves}  |  Пары: #{pairs_found}/#{total_pairs}", :bold)
    display_board(board, revealed, size)

    r2, c2 = get_coordinates(size)
    if (r2 == r1 && c2 == c1) || revealed[r2][c2]
      puts colorize("Неверный выбор.", :yellow)
      revealed[r1][c1] = false
      next
    end
    revealed[r2][c2] = true
    moves += 1
    clear_screen
    puts colorize("🧠  МЕМОРИ  |  Размер #{size}×#{size}  |  Ходы: #{moves}  |  Пары: #{pairs_found}/#{total_pairs}", :bold)
    display_board(board, revealed, size)

    if board[r1][c1] == board[r2][c2]
      puts colorize("✅ Пара найдена!", :green)
      pairs_found += 1
      sleep 1
    else
      puts colorize("❌ Не совпало.", :red)
      sleep timeout > 0 ? timeout : 1.5
      revealed[r1][c1] = false
      revealed[r2][c2] = false
    end
  end

  elapsed = (Time.now - start_time).to_i
  puts colorize("\n🎉 Поздравляем! Вы завершили игру за #{moves} ходов и #{elapsed} секунд.", :green)

  best = records[key]
  if best.nil? || moves < best['moves'] || (moves == best['moves'] && elapsed < best['time'])
    records[key] = { 'moves' => moves, 'time' => elapsed, 'date' => Time.now.iso8601 }
    save_records(records)
    puts colorize("🏆 Новый рекорд для размера #{size}×#{size}!", :yellow)
  else
    puts colorize("Лучший результат для этого размера: #{best['moves']} ходов за #{best['time']} сек.", :blue)
  end
end

main if __FILE__ == $0
