package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const bufferSize = 5           // Размер буфера
const bufferFlushInterval = 10 // Интервал опустошения буфера (в секундах)

// Функция, выполняющая фильтрацию отрицательных чисел
func filterNegativeNumbers(input <-chan int) <-chan int {
	log.Println("The beginning of the negative number filtering stage")
	output := make(chan int)
	go func() {
		for num := range input {
			if num >= 0 {
				output <- num
			}
		}
		close(output)
	}()
	return output
}

// Функция, выполняющая фильтрацию чисел, не кратных 3
func filterNonMultipleOfThree(input <-chan int) <-chan int {
	log.Println("The beginning of the positive number filtering stage")
	output := make(chan int)
	go func() {
		for num := range input {
			if num%3 != 0 && num != 0 {
				output <- num
			}
		}
		close(output)
	}()
	return output
}

// Структура кольцевого буфера
type RingBuffer struct {
	buffer      []int
	size        int
	writeCursor int
	readCursor  int
}

// Инициализация кольцевого буфера
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer:      make([]int, size),
		size:        size,
		writeCursor: 0,
		readCursor:  0,
	}
}

// Метод добавления элемента в буфер
func (r *RingBuffer) Add(item int) {
	r.buffer[r.writeCursor] = item
	r.writeCursor = (r.writeCursor + 1) % r.size
	if r.writeCursor == r.readCursor {
		r.readCursor = (r.readCursor + 1) % r.size
	}
}

// Метод опустошения буфера
func (r *RingBuffer) Flush() []int {
	items := make([]int, 0)
	for r.readCursor != r.writeCursor {
		items = append(items, r.buffer[r.readCursor])
		r.readCursor = (r.readCursor + 1) % r.size
	}
	return items
}

// Функция, выполняющая буферизацию данных в кольцевом буфере
func bufferData(input <-chan int, bufferSize int, flushInterval time.Duration) <-chan []int {
	output := make(chan []int)
	buffer := NewRingBuffer(bufferSize)
	go func() {
		ticker := time.NewTicker(flushInterval * time.Second)
		for {
			select {
			case num, ok := <-input:
				if ok {
					log.Printf("Add a number %d in the buffer", num)
					buffer.Add(num)
				} else {
					ticker.Stop()
					items := buffer.Flush()
					if len(items) > 0 {
						output <- items
					}
					close(output)
					return
				}
			case <-ticker.C:
				items := buffer.Flush()
				log.Println("Buffer is empty")
				if len(items) > 0 {
					output <- items
				}
			}
		}
	}()
	return output
}

// Источник данных для конвейера
func dataSource(output chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		num, err := strconv.Atoi(text)
		log.Printf("Get a data from data source")
		if err == nil {
			output <- num
		} else if strings.EqualFold(text, "exit") {
			fmt.Println("Программа закончена")
			log.Println("The pipline was stopped")
			os.Exit(0)

		} else {
			fmt.Printf("Неопознанная команда %v ,введите число: ", text)
		}

	}
	close(output)
}

// Потребитель данных конвейера
func dataConsumer(input <-chan []int, wg *sync.WaitGroup) {
	defer wg.Done()
	for items := range input {
		log.Println("Get a data from buffer")
		fmt.Printf("Получены данные:%v\n", items)
	}
}
func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	log.Println("The Pipline was started")

	sourceOutput := make(chan int)
	filter1Output := filterNegativeNumbers(sourceOutput)
	filter2Output := filterNonMultipleOfThree(filter1Output)
	bufferOutput := bufferData(filter2Output, bufferSize, bufferFlushInterval)

	go dataConsumer(bufferOutput, &wg)
	go dataSource(sourceOutput, &wg)

	wg.Wait()
}
