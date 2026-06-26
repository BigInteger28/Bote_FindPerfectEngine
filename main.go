package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// **elementsDepthArray** en **moveToIndexArray**: arrays ipv maps voor snelle hot-path lookups
var elementsDepthArray [256][4]byte
var moveToIndexArray [256]int

func init() {
	elementsDepthArray['W'] = [4]byte{'L', 'A', 'V', 'W'}
	elementsDepthArray['V'] = [4]byte{'W', 'L', 'A', 'V'}
	elementsDepthArray['A'] = [4]byte{'V', 'W', 'L', 'A'}
	elementsDepthArray['L'] = [4]byte{'A', 'V', 'W', 'L'}
	moveToIndexArray['W'] = 0
	moveToIndexArray['V'] = 1
	moveToIndexArray['A'] = 2
	moveToIndexArray['L'] = 3
	moveToIndexArray['D'] = 4
}

// **moveWins** definieert wie wint (1 = move1 wint, 2 = move2 wint, 0 = gelijk)
var moveWins = [5][5]uint8{
	{0, 1, 0, 2, 0}, // W vs W,V,A,L,D
	{2, 0, 1, 0, 0}, // V
	{0, 2, 0, 1, 0}, // A
	{1, 0, 2, 0, 0}, // L
	{0, 0, 0, 0, 0}, // D
}

// **printMu** serialiseert prints over goroutines
var printMu sync.Mutex

// **depthToElement** converteert een diepte naar een element (alleen dieptes 1-5)
var depthToElement = [5]byte{'W', 'V', 'A', 'L', 'D'}

// **Player** houdt de staat van een speler bij
type Player struct {
	available [5]int // W, V, A, L, D
	moves     [13]byte
	moveCount int
}

// **engineResult** houdt een engine en zijn totaalscore bij
type engineResult struct {
	engine string
	score  int
}

// **minHeap** implementeert heap.Interface voor de top 10.000 engines (hoogste scores eerst)
type minHeap []engineResult

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].score < h[j].score } // Min-heap, laagste score eerst
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(engineResult))
}
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Globale variabelen voor voortgang
var progressComparisons int64
var totalComparisons int64
var updateInterval int64 = 10000000 // Update na elke 10.000.000 matches, aanpasbaar
var startTime time.Time

// **getElementFromCode** haalt direct een element op basis van de engine code voor de eerste zet
func getElementFromCode(depth int) byte {
	if depth < 1 || depth > 5 {
		return 0
	}
	return depthToElement[depth-1]
}

// **getElementByDepth** berekent het volgende element gebaseerd op vorig element en diepte
func getElementByDepth(prevElement byte, depth int) byte {
	if depth == 5 {
		return 'D'
	}
	if prevElement == 0 {
		return 0
	}
	if prevElement == 'D' {
		prevElement = 'L'
	}
	return elementsDepthArray[prevElement][depth-1]
}

// **chooseAvailableElement** kiest een beschikbaar element of alternatief met diepte 1 fallback
func chooseAvailableElement(target byte, available *[5]int) byte {
	targetIdx := moveToIndexArray[target]
	if available[targetIdx] > 0 {
		return target
	}
	current := target
	for i := 0; i < 5; i++ {
		current = elementsDepthArray[current][0] // depth 1 = index 0
		currentIdx := moveToIndexArray[current]
		if available[currentIdx] > 0 {
			return current
		}
	}
	if available[4] > 0 { // D
		return 'D'
	}
	return 0
}

// **getLastElement** bepaalt het resterende element voor de 13e zet
func getLastElement(available *[5]int) byte {
	for i, c := range depthToElement {
		if available[i] > 0 {
			return c
		}
	}
	return 0
}

// **determineWinner** bepaalt de winnaar met bitwise operaties
func determineWinner(move1, move2 byte) int {
	return int(moveWins[moveToIndexArray[move1]][moveToIndexArray[move2]])
}

// **simulateDepthGame** simuleert een spel met diepte-gebaseerde codes
func simulateDepthGame(engine1, engine2 string) (p1Score, p2Score int) {
	if len(engine1) != 12 || len(engine2) != 12 {
		return -1, -1
	}

	var p1, p2 Player
	p1.available = [5]int{3, 3, 3, 3, 1}
	p2.available = [5]int{3, 3, 3, 3, 1}

	for i := 0; i < 12; i++ {
		depth1 := int(engine1[i] - '0')
		depth2 := int(engine2[i] - '0')

		var target1, target2 byte
		if depth1 >= 6 {
			base1 := depth1 - 5
			if i < 2 {
				target1 = getElementFromCode(base1)
			} else {
				target1 = getElementByDepth(p2.moves[i-2], base1)
			}
		} else if i == 0 {
			target1 = getElementFromCode(depth1)
		} else {
			target1 = getElementByDepth(p2.moves[i-1], depth1)
		}
		if depth2 >= 6 {
			base2 := depth2 - 5
			if i < 2 {
				target2 = getElementFromCode(base2)
			} else {
				target2 = getElementByDepth(p1.moves[i-2], base2)
			}
		} else if i == 0 {
			target2 = getElementFromCode(depth2)
		} else {
			target2 = getElementByDepth(p1.moves[i-1], depth2)
		}
		move1 := chooseAvailableElement(target1, &p1.available)
		move2 := chooseAvailableElement(target2, &p2.available)

		if move1 == 0 || move2 == 0 {
			return -1, -1
		}

		p1.available[moveToIndex[move1]]--
		p1.moves[p1.moveCount] = move1
		p1.moveCount++
		p2.available[moveToIndex[move2]]--
		p2.moves[p2.moveCount] = move2
		p2.moveCount++

		winner := determineWinner(move1, move2)
		if winner == 1 {
			p1Score++
		} else if winner == 2 {
			p2Score++
		}

		// Early termination: als p1 niet meer kan winnen of gelijkspelen
		if p2Score-p1Score > 12-i {
			return p1Score, p2Score
		}
	}

	move1 := getLastElement(&p1.available)
	move2 := getLastElement(&p2.available)
	if move1 != 0 {
		p1.available[moveToIndex[move1]]--
		p1.moves[p1.moveCount] = move1
		p1.moveCount++
	}
	if move2 != 0 {
		p2.available[moveToIndex[move2]]--
		p2.moves[p2.moveCount] = move2
		p2.moveCount++
	}

	winner := determineWinner(move1, move2)
	if winner == 1 {
		p1Score++
	} else if winner == 2 {
			p2Score++
		}

	return p1Score, p2Score
}

// **simulateFixedGame** simuleert een spel met vaste zetten
func simulateFixedGame(engine1, engine2 string) (p1Score, p2Score int) {
	if len(engine1) != 13 || len(engine2) != 13 {
		return -1, -1
	}

	for i := 0; i < 13; i++ {
		move1, move2 := engine1[i], engine2[i]
		winner := determineWinner(move1, move2)
		if winner == 1 {
			p1Score++
		} else if winner == 2 {
			p2Score++
		}
	}

	return p1Score, p2Score
}

// **generateEngines** genereert alle engine codes met max 1 '5', alle dieptes 1-5
func generateEngines(startDepth string) []string {
	var engines []string
	remainingLength := 12 - len(startDepth)
	hasFive := strings.Contains(startDepth, "5")

	if remainingLength < 0 {
		return engines
	}

	if startDepth != "" {
		for i, digit := range startDepth {
			if digit < '1' || digit > '9' {
				return engines
			}
			if i == 0 && digit > '5' {
				return engines // eerste positie mag alleen 1-5 zijn
			}
		}
		generateRemaining(startDepth, remainingLength, hasFive, &engines)
	} else {
		for firstDigit := '1'; firstDigit <= '5'; firstDigit++ { // eerste positie: enkel 1-5
			prefix := string(firstDigit)
			hasFiveLocal := firstDigit == '5'
			generateRemaining(prefix, 11, hasFiveLocal, &engines)
		}
	}

	return engines
}

// **generateRemaining** genereert de resterende posities iteratief
func generateRemaining(prefix string, remainingLength int, hasUsedFive bool, engines *[]string) {
	if remainingLength == 0 {
		if len(prefix) == 12 {
			*engines = append(*engines, prefix)
		}
		return
	}

	for digit := '1'; digit <= '9'; digit++ {
		if digit == '5' && hasUsedFive {
			continue
		}
		newPrefix := prefix + string(digit)
		generateRemaining(newPrefix, remainingLength-1, hasUsedFive || digit == '5', engines)
	}
}

// **simulateDepthGameToMoves** genereert de zetten van een diepte-gebaseerde engine, reactief op de tegenstander
func simulateDepthGameToMoves(engine string, opponent string) (moves [13]byte) {
	if len(engine) != 12 || len(opponent) != 13 {
		return
	}

	p := Player{
		available: [5]int{3, 3, 3, 3, 1},
	}

	for i := 0; i < 12; i++ {
		depth := int(engine[i] - '0')
		var target byte
		if depth >= 6 {
			base := depth - 5
			if i < 2 {
				target = getElementFromCode(base)
			} else {
				target = getElementByDepth(opponent[i-2], base)
			}
		} else if i == 0 {
			target = getElementFromCode(depth)
		} else {
			target = getElementByDepth(opponent[i-1], depth)
		}
		move := chooseAvailableElement(target, &p.available)
		if move != 0 {
			p.available[moveToIndexArray[move]]--
			moves[p.moveCount] = move
			p.moveCount++
		} else {
			move = getLastElement(&p.available)
			if move != 0 {
				p.available[moveToIndexArray[move]]--
				moves[p.moveCount] = move
				p.moveCount++
			}
		}
	}

	move := getLastElement(&p.available)
	if move != 0 {
		p.available[moveToIndexArray[move]]--
		moves[p.moveCount] = move
	} else {
		moves[p.moveCount] = 'W'
	}

	return moves
}

// **evaluateBatch** evalueert een batch van engines en berekent de totale score met bonus/malus
func evaluateBatch(engines []string, inputEngines []string, top10000Chan chan<- engineResult, progressComparisons *int64) {
    h := &minHeap{}
    heap.Init(h)
    maxSize := 10000

    for _, engine := range engines {
        totalScore := 0
        for _, inputEngine := range inputEngines {
            var p1Score, p2Score int
            if len(inputEngine) == 13 {
                if len(engine) == 12 {
                    p1Moves := simulateDepthGameToMoves(engine, inputEngine)
                    p1Score, p2Score = simulateFixedGame(string(p1Moves[:]), inputEngine)
                } else {
                    p1Score, p2Score = simulateFixedGame(engine, inputEngine)
                }
            } else {
                p1Score, p2Score = simulateDepthGame(engine, inputEngine)
            }
            if p1Score == -1 || p2Score == -1 {
                continue
            }

            // Nieuwe scoreberekening
            diff := p1Score - p2Score
            if p1Score > p2Score {
                totalScore += diff + 10 // Winst: +10 bonus
            } else if p1Score < p2Score {
                totalScore += diff - 10 // Verlies: -10 malus
            } else {
                totalScore += p1Score   // Gelijkspel: +p1Score
            }

            // Update voortgang na elke match
            atomic.AddInt64(progressComparisons, 1)
            p := atomic.LoadInt64(progressComparisons)
            if p % updateInterval == 0 {
                printMu.Lock()
                elapsed := time.Since(startTime).Seconds()
                if elapsed > 0 {
                    speed := float64(p) / elapsed / 1000 // k matches/s
                    fmt.Printf("Voortgang: %d / %d matches (%.2f%%), Snelheid: %.1f k matches/s\n",
                        p, totalComparisons, float64(p)/float64(totalComparisons)*100, speed)
                } else {
                    fmt.Printf("Voortgang: %d / %d matches (%.2f%%)\n",
                        p, totalComparisons, float64(p)/float64(totalComparisons)*100)
                }
                printMu.Unlock()
            }
        }

        result := engineResult{engine: engine, score: totalScore}
        if h.Len() < maxSize {
            heap.Push(h, result)
        } else if totalScore > (*h)[0].score {
            heap.Pop(h)
            heap.Push(h, result)
        }
    }

    for h.Len() > 0 {
        top10000Chan <- heap.Pop(h).(engineResult)
    }
}

// **parseEngineCode** haalt de engine code uit een invoer met prefix en kapt af na 12 cijfers voor depth engines
func parseEngineCode(input string) string {
	parts := strings.Split(input, ":")
	engine := strings.TrimSpace(input)
	if len(parts) > 2 {
		engine = strings.TrimSpace(parts[2])
	}
	if len(engine) > 12 && strings.ContainsAny(engine, "123456789") && !strings.ContainsAny(engine, "WVALD") {
		return engine[:12]
	}
	return engine
}

func main() {
	for {
		var inputEngines []string
		fmt.Println("Voer engine codes in (één per regel, '.' om te stoppen):")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if input == "." || input == "" {
				break
			}
			engine := parseEngineCode(input)
			validDepth := len(engine) == 12 && !strings.ContainsAny(engine, "0") && strings.ContainsAny(engine, "123456789")
			validFixed := len(engine) == 13 && strings.ContainsAny(engine, "WVALD") && !strings.ContainsAny(engine, "1234567890")
			if validDepth || validFixed {
				inputEngines = append(inputEngines, engine)
			} else {
				fmt.Printf("Ongeldige engine code '%s'. Moet 12 chiffres (1-9) of 13 tekens (W, V, A, L, D) zijn.\n", engine)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Fout bij het lezen van invoer: %v\n", err)
			continue
		}

		if len(inputEngines) == 0 {
			fmt.Println("Geen engine codes ingevoerd. Gestopt.")
			break
		}

		var startDepth string
		var maxMemoryMB int

		fmt.Println("Voer de startdepth in (leeg voor alle combinaties, bijv. '51'): ")
		fmt.Scanln(&startDepth)

		if len(startDepth) > 12 || (startDepth != "" && strings.ContainsAny(startDepth, "0")) {
			fmt.Println("Ongeldige startdepth. Moet <= 12 chiffres zijn, eerste positie 1-5, rest 1-9.")
			continue
		}

		fmt.Println("Voer het maximale geheugen in MB in (1-512000, default 128000): ")
		var memoryInput string
		fmt.Scanln(&memoryInput)
		if memoryInput == "" {
			maxMemoryMB = 64000
		} else if n, err := fmt.Sscanf(memoryInput, "%d", &maxMemoryMB); err != nil || n != 1 || maxMemoryMB < 1 || maxMemoryMB > 512000 {
			maxMemoryMB = 64000
			fmt.Println("Ongeldige invoer, defaulting naar 64.000 MB.")
		}

		generatedEngines := generateEngines(startDepth)

		const bytesPerResult = 24
		maxBufferSize := (maxMemoryMB * 1024 * 1024) / bytesPerResult
		if maxBufferSize > len(generatedEngines) {
			maxBufferSize = len(generatedEngines)
		}
		if maxBufferSize < 10000 {
			maxBufferSize = 10000
		}

		totalEngines := len(generatedEngines)
		numInputEngines := len(inputEngines)
		totalComparisons = int64(totalEngines) * int64(numInputEngines)

		top10000Chan := make(chan engineResult, 1000000)
		var wg sync.WaitGroup
		progressComparisons = 0
		startTime = time.Now()

		defaultThreads := runtime.NumCPU()
		fmt.Printf("Voer het aantal threads in (default %d): ", defaultThreads)
		var threadsInput string
		fmt.Scanln(&threadsInput)
		numThreads := defaultThreads
		if threadsInput != "" {
			if n, err := fmt.Sscanf(threadsInput, "%d", &numThreads); err != nil || n != 1 || numThreads < 1 {
				numThreads = defaultThreads
				fmt.Printf("Ongeldige invoer, defaulting naar %d threads.\n", defaultThreads)
			}
		}
		enginesPerThread := (totalEngines + numThreads - 1) / numThreads

		for i := 0; i < numThreads; i++ {
			start := i * enginesPerThread
			end := start + enginesPerThread
			if end > totalEngines {
				end = totalEngines
			}
			wg.Add(1)
			go func(threadStart, threadEnd int) {
				defer wg.Done()
				batch := generatedEngines[threadStart:threadEnd]
				evaluateBatch(batch, inputEngines, top10000Chan, &progressComparisons)
			}(start, end)
		}

		go func() {
			wg.Wait()
			close(top10000Chan)
		}()

		file, err := os.Create("top_10000_engines.txt")
		if err != nil {
			fmt.Printf("Fout bij het openen van bestand: %v\n", err)
			return
		}
		defer file.Close()

		top10000 := &minHeap{}
		heap.Init(top10000)
		maxSize := 10000
		for result := range top10000Chan {
			if top10000.Len() < maxSize {
				heap.Push(top10000, result)
			} else if result.score > (*top10000)[0].score {
				heap.Pop(top10000)
				heap.Push(top10000, result)
			}
		}

		if top10000.Len() > 0 {
			results := make([]engineResult, 0, maxSize)
			for top10000.Len() > 0 {
				results = append(results, heap.Pop(top10000).(engineResult))
			}
			for i := len(results) - 1; i >= 0; i-- {
				result := results[i]
				_, err := file.WriteString(fmt.Sprintf("%s (score: %d)\n", result.engine, result.score))
				if err != nil {
					fmt.Printf("Fout bij het schrijven: %v\n", err)
					break
				}
			}
			fmt.Printf("Top 10.000 engines opgeslagen uit %d matches.\n", totalComparisons)
		} else {
			fmt.Println("Geen engines geëvalueerd.")
		}
	}
	fmt.Println("Gestopt.")
}
