package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// elementsDepth definieert de transformaties van elementen op verschillende dieptes
var elementsDepth = map[byte]map[int]byte{
	'W': {1: 'L', 2: 'A', 3: 'V', 4: 'W'},
	'V': {1: 'W', 2: 'L', 3: 'A', 4: 'V'},
	'A': {1: 'V', 2: 'W', 3: 'L', 4: 'A'},
	'L': {1: 'A', 2: 'V', 3: 'W', 4: 'L'},
}

// moveWins definieert bitwise wie wint (1 = move1 wint, 2 = move2 wint, 0 = gelijk)
var moveWins = [5][5]uint8{
	{0, 1, 0, 2, 0}, // W vs W,V,A,L,D
	{2, 0, 1, 0, 0}, // V
	{0, 2, 0, 1, 0}, // A
	{1, 0, 2, 0, 0}, // L
	{0, 0, 0, 0, 0}, // D
}

// moveToIndex converteert een move naar een index
var moveToIndex = map[byte]int{
	'W': 0,
	'V': 1,
	'A': 2,
	'L': 3,
	'D': 4,
}

// depthToElement converteert een diepte naar een element (alleen dieptes 1-5)
var depthToElement = [5]byte{'W', 'V', 'A', 'L', 'D'}

// Player houdt de staat van een speler bij
type Player struct {
	available [5]int // W, V, A, L, D
	moves     [13]byte
	moveCount int
}

// engineResult houdt een engine en zijn totaalscore bij
type engineResult struct {
	engine string
	score  int
}

// minHeap implementeert heap.Interface voor de top 10,000 engines (hoogste scores eerst)
type minHeap []engineResult

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].score < h[j].score } // Max-heap op positieve score (hoogste scores eerst)
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

// getElementFromCode haalt direct een element op basis van de engine code voor de eerste zet
func getElementFromCode(depth int) byte {
	if depth < 1 || depth > 5 {
		return 0 // Ongeldige diepte, retourneer 0 (ongeldige move)
	}
	if depth == 5 {
		return 'D' // Speciaal geval voor diepte 5
	}
	return depthToElement[depth-1] // Offset met 1 omdat depthToElement nu 1-4 is voor 'W', 'V', 'A', 'L'
}

// getElementByDepth berekent het volgende element gebaseerd op vorig element en diepte
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
	next, ok := elementsDepth[prevElement][depth]
	if !ok {
		return 0
	}
	return next
}

// chooseAvailableElement kiest een beschikbaar element of alternatief met diepte 1 fallback
func chooseAvailableElement(target byte, available *[5]int) byte {
	targetIdx := moveToIndex[target]
	if available[targetIdx] > 0 {
		return target
	}
	current := target
	for i := 0; i < 5; i++ {
		current = elementsDepth[current][1]
		currentIdx := moveToIndex[current]
		if available[currentIdx] > 0 {
			return current
		}
	}
	if available[4] > 0 { // D
		return 'D'
	}
	return 0
}

// getLastElement bepaalt het resterende element voor de 13e zet
func getLastElement(available *[5]int) byte {
	candidates := [5]byte{'W', 'V', 'A', 'L', 'D'}
	for _, c := range candidates {
		if available[moveToIndex[c]] > 0 {
			return c
		}
	}
	return 0
}

// determineWinner bepaalt de winnaar met bitwise operaties
func determineWinner(move1, move2 byte) int {
	move1Idx, ok1 := moveToIndex[move1]
	move2Idx, ok2 := moveToIndex[move2]
	if !ok1 || !ok2 {
		return 0 // Ongeldige moves, geen winnaar
	}
	return int(moveWins[move1Idx][move2Idx])
}

// simulateDepthGame simuleert een spel met diepte-gebaseerde codes
func simulateDepthGame(engine1, engine2 string) (p1Score, p2Score int) {
	if len(engine1) != 12 || len(engine2) != 12 {
		return -1, -1
	}

	var p1, p2 Player
	p1.available = [5]int{3, 3, 3, 3, 1}
	p2.available = [5]int{3, 3, 3, 3, 1}

	p1Score, p2Score = 0, 0

	for i := 0; i < 12; i++ {
		depth1, depth2 := int(engine1[i]-'0'), int(engine2[i]-'0')

		var move1, move2 byte
		if i == 0 {
			move1 = chooseAvailableElement(getElementFromCode(depth1), &p1.available)
			move2 = chooseAvailableElement(getElementFromCode(depth2), &p2.available)
		} else {
			move1 = chooseAvailableElement(getElementByDepth(p2.moves[i-1], depth1), &p1.available)
			move2 = chooseAvailableElement(getElementByDepth(p1.moves[i-1], depth2), &p2.available)
		}

		if move1 == 0 || move2 == 0 {
			return -1, -1 // Ongeldige move, skip deze engine
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

// simulateFixedGame simuleert een spel met vaste zetten
func simulateFixedGame(engine1, engine2 string) (p1Score, p2Score int) {
	if len(engine1) != 13 || len(engine2) != 13 {
		return -1, -1
	}

	p1Score, p2Score = 0, 0

	for i := 0; i < 13; i++ {
		move1, move2 := engine1[i], engine2[i]
		validMoves := map[byte]bool{'W': true, 'V': true, 'A': true, 'L': true, 'D': true}
		if !validMoves[move1] || !validMoves[move2] {
			return -1, -1 // Ongeldige move, skip deze engine
		}
		winner := determineWinner(move1, move2)
		if winner == 1 {
			p1Score++
		} else if winner == 2 {
			p2Score++
		}
	}

	return p1Score, p2Score
}

// generateEngines genereert alle engine codes met max 1 '5', alle dieptes 1-5
func generateEngines(startDepth string) []string {
	var engines []string
	remainingLength := 12 - len(startDepth)
	hasFive := strings.Contains(startDepth, "5")

	if remainingLength < 0 {
		return engines
	}

	if startDepth != "" {
		// Valideer startDepth (alleen dieptes 1-5)
		for _, digit := range startDepth {
			if digit < '1' || digit > '5' {
				return engines // Ongeldige startdepth, retourneer lege lijst
			}
		}
		if len(startDepth) > 12 {
			return engines // Ongeldige startdepth, retourneer lege lijst
		}
		generateRemaining(startDepth, remainingLength, hasFive, &engines)
	} else {
		// Genereer alle engines van 12 posities, beginnend met 1-5
		for firstDigit := '1'; firstDigit <= '5'; firstDigit++ {
			prefix := string(firstDigit)
			hasFiveLocal := firstDigit == '5'
			generateRemaining(prefix, 11, hasFiveLocal, &engines)
		}
	}

	return engines
}

// generateRemaining genereert de resterende posities iteratief
func generateRemaining(prefix string, remainingLength int, hasUsedFive bool, engines *[]string) {
	if remainingLength == 0 {
		if len(prefix) == 12 && (prefix[11] == '1' || prefix[11] == '3') {
			*engines = append(*engines, prefix)
		}
		return
	}

	if remainingLength == 1 { // Laatste positie, forceer 1 of 3
		generateRemaining(prefix+"1", remainingLength-1, hasUsedFive, engines)
		generateRemaining(prefix+"3", remainingLength-1, hasUsedFive, engines)
		return
	}

	// Normale posities (niet de laatste)
	for digit := '1'; digit <= '4'; digit++ {
		newPrefix := prefix + string(digit)
		generateRemaining(newPrefix, remainingLength-1, hasUsedFive, engines)
	}
	if !hasUsedFive {
		newPrefix := prefix + "5"
		generateRemaining(newPrefix, remainingLength-1, true, engines)
	}
}

// simulateDepthGameToMoves genereert de zetten van een diepte-gebaseerde engine, reactief op de tegenstander
func simulateDepthGameToMoves(engine string, opponent string) (moves [13]byte) {
	if len(engine) != 12 || len(opponent) != 13 {
		return
	}

	p := Player{
		available: [5]int{3, 3, 3, 3, 1},
	}

	for i := 0; i < 12; i++ {
		depth := int(engine[i] - '0')
		var move byte
		if i == 0 {
			move = chooseAvailableElement(getElementFromCode(depth), &p.available)
		} else {
			move = chooseAvailableElement(getElementByDepth(opponent[i-1], depth), &p.available)
		}
		if move != 0 {
			p.available[moveToIndex[move]]--
			moves[p.moveCount] = move
			p.moveCount++
		} else {
			move = getLastElement(&p.available)
			if move != 0 {
				p.available[moveToIndex[move]]--
				moves[p.moveCount] = move
				p.moveCount++
			}
		}
	}

	move := getLastElement(&p.available)
	if move != 0 {
		p.available[moveToIndex[move]]--
		moves[p.moveCount] = move
	} else {
		moves[p.moveCount] = 'W'
	}

	return moves
}

// evaluateBatch evalueert een batch van engines en berekent de totale score (p1Score - p2Score)
func evaluateBatch(engines []string, inputEngines []string, top10000Chan chan<- engineResult, progress *int32) {
	h := &minHeap{} // Max-heap voor hoogste scores eerst
	heap.Init(h)
	maxSize := 10000

	for _, engine := range engines {
		totalScore := 0 // p1Score - p2Score over all games

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
				continue // Skip invalid engines
			}
			totalScore += p1Score - p2Score // Score is difference between wins and losses
		}

		if totalScore != 0 { // Alleen engines met een score > 0 toevoegen
			result := engineResult{engine: engine, score: totalScore}
			if h.Len() < maxSize {
				heap.Push(h, result)
			} else if totalScore > (*h)[0].score {
				heap.Pop(h)
				heap.Push(h, result)
			}
		}
		atomic.AddInt32(progress, int32(len(inputEngines))) // Tel aantal inputEngines per engine
	}

	top10000 := make([]engineResult, 0, maxSize)
	for h.Len() > 0 {
		top10000 = append(top10000, heap.Pop(h).(engineResult))
	}

	for _, result := range top10000 {
		top10000Chan <- result
	}
}

// parseEngineCode haalt de engine code uit een invoer met prefix en kapt af na 12 cijfers voor depth engines
func parseEngineCode(input string) string {
	parts := strings.Split(input, ":")
	engine := strings.TrimSpace(input)
	if len(parts) > 2 {
		engine = strings.TrimSpace(parts[2])
	}
	// Als het een depth engine is (alleen cijfers), kap af na 12 cijfers
	if len(engine) > 12 && strings.ContainsAny(engine, "12345") && !strings.ContainsAny(engine, "WVALD") {
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
			validDepth := len(engine) == 12 && !strings.ContainsAny(engine, "67890") && strings.ContainsAny(engine, "12345")
			validFixed := len(engine) == 13 && strings.ContainsAny(engine, "WVALD") && !strings.ContainsAny(engine, "1234567890")
			if validDepth || validFixed {
				inputEngines = append(inputEngines, engine)
			} else {
				fmt.Printf("Ongeldige engine code '%s'. Moet 12 chiffres (1-5) of 13 tekens (W, V, A, L, D) zijn.\n", engine)
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

		if len(startDepth) > 12 || (startDepth != "" && strings.ContainsAny(startDepth, "67890")) {
			fmt.Println("Ongeldige startdepth. Moet <= 12 chiffres zijn, alleen 1-5 of leeg.")
			continue
		}

		// Myeloid RAM-instelling
		fmt.Println("Voer het maximale geheugen in MB in (1-512000, default 128000): ")
		var memoryInput string
		fmt.Scanln(&memoryInput)
		if memoryInput == "" {
			maxMemoryMB = 128000 // Standaardwaarde
		} else if n, err := fmt.Sscanf(memoryInput, "%d", &maxMemoryMB); err != nil || n != 1 || maxMemoryMB < 1 || maxMemoryMB > 512000 {
			maxMemoryMB = 512000 // Default naar 128,000 MB als invoer ongeldig
			fmt.Println("Ongeldige invoer, defaulting naar 128,000 MB.")
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

		// Pre-allocate channel for top 10,000 results
		top10000Chan := make(chan engineResult, 1000000) // Grotere buffer voor maximale snelheid
		var wg sync.WaitGroup
		var progress int32

		// Totaal aantal te evalueren combinaties: generatedEngines * inputEngines
		totalEngines := len(generatedEngines) * len(inputEngines)

		var startTime time.Time
		go func(total int) {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case t := <-ticker.C:
					p := atomic.LoadInt32(&progress)
					if p > 0 && p <= int32(total) {
						speed := float64(p) / t.Sub(startTime).Seconds() / 1000 // Delen door 1000 voor k engines/s
						fmt.Printf("Progress: %d / %d engines (%.2f%%), Speed: %.1f k engines/s\n", 
							p, total, float64(p)/float64(total)*100, speed)
					}
				case <-time.After(30 * time.Second): // Stop na 30 seconden als er geen voortgang is
					return
				}
			}
		}(totalEngines)

		startTime = time.Now()
		const numThreads = 64 // Standaard op 64 threads
		enginesPerThread := (len(generatedEngines) + numThreads - 1) / numThreads // Gelijkmatige verdeling

		for i := 0; i < numThreads; i++ {
			start := i * enginesPerThread
			end := start + enginesPerThread
			if end > len(generatedEngines) {
				end = len(generatedEngines)
			}

			wg.Add(1)
			go func(threadStart, threadEnd int) {
				defer wg.Done()
				batch := generatedEngines[threadStart:threadEnd]
				evaluateBatch(batch, inputEngines, top10000Chan, &progress)
			}(start, end)
		}

		// Goroutine om top 10,000 te schrijven naar bestand
		go func() {
			wg.Wait()
			close(top10000Chan)
		}()

		// Schrijf top 10,000 direct naar bestand met max-heap
		file, err := os.Create("top_10000_engines.txt")
		if err != nil {
			fmt.Printf("Fout bij het openen van bestand: %v\n", err)
			return
		}
		defer file.Close()

		top10000 := &minHeap{}
		heap.Init(top10000)
		maxSize := 10000
		count := 0

		for result := range top10000Chan {
			count++
			if top10000.Len() < maxSize {
				heap.Push(top10000, result)
			} else if result.score > (*top10000)[0].score {
				heap.Pop(top10000)
				heap.Push(top10000, result)
			}
			if count%10000 == 0 { // Update progress per 10,000 engines
				atomic.AddInt32(&progress, 10000)
			}
		}

		if top10000.Len() > 0 {
			// Verzamel en schrijf top 10,000 in volgorde (hoogste naar laagste score)
			results := make([]engineResult, 0, maxSize)
			for top10000.Len() > 0 {
				results = append(results, heap.Pop(top10000).(engineResult))
			}
			// Schrijf van achter naar voren voor hoogste naar laagste
			for i := len(results)-1; i >= 0; i-- {
				result := results[i]
				_, err := file.WriteString(fmt.Sprintf("%s (score: %d)\n", result.engine, result.score))
				if err != nil {
					fmt.Printf("Fout bij het schrijven naar bestand: %v\n", err)
					break
				}
			}
			fmt.Printf("We found the top 10,000 engines / %d total generated engines, sorted by highest score (most wins, fewest losses)\n",
				len(generatedEngines))
		} else {
			fmt.Println("No engines evaluated due to errors. Check for invalid moves or logic errors.")
		}
	}
	fmt.Println("Gestopt.")
}
