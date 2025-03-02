package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
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
	engineCode string
	available  [5]int // W, V, A, L, D
	moves      [13]byte
	moveCount  int
}

// engineResult houdt een engine en zijn totaalscore bij
type engineResult struct {
	engine string
	score  int
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
	return elementsDepth[prevElement][depth]
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
	return int(moveWins[moveToIndex[move1]][moveToIndex[move2]])
}

// simulateDepthGame simuleert een spel met diepte-gebaseerde codes
func simulateDepthGame(engine1, engine2 string) (result int, p1Score, p2Score int) {
	if len(engine1) != 12 || len(engine2) != 12 {
		return -1, 0, 0
	}

	var p1, p2 Player
	p1.available = [5]int{3, 3, 3, 3, 1}
	p2.available = [5]int{3, 3, 3, 3, 1}
	p1.engineCode, p2.engineCode = engine1, engine2

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

	if p1Score > p2Score {
		return 1, p1Score, p2Score
	} else if p2Score > p1Score {
		return 2, p1Score, p2Score
	}
	return 0, p1Score, p2Score
}

// simulateFixedGame simuleert een spel met vaste zetten
func simulateFixedGame(engine1, engine2 string) (result int, p1Score, p2Score int) {
	if len(engine1) != 13 || len(engine2) != 13 {
		return -1, 0, 0
	}

	p1Score, p2Score = 0, 0

	for i := 0; i < 13; i++ {
		move1, move2 := engine1[i], engine2[i]
		winner := determineWinner(move1, move2)
		if winner == 1 {
			p1Score++
		} else if winner == 2 {
			p2Score++
		}
	}

	if p1Score > p2Score {
		return 1, p1Score, p2Score
	} else if p2Score > p1Score {
		return 2, p1Score, p2Score
	}
	return 0, p1Score, p2Score
}

// generateWithLastDigit genereert een engine, van voor naar achter, max 1 '5' (alle laatste cijfers 1-5)
func generateWithLastDigit(prefix string, length int, hasUsedFive bool, engines *[]string) {
	if length == 0 {
		if len(prefix) == 12 {
			// Debug: Controleer of we engines zoals 331211513334 genereren
			// fmt.Printf("Generated engine: %s\n", prefix)
			*engines = append(*engines, prefix)
		}
		return
	}
	for digit := '5'; digit >= '1'; digit-- { // Genereren van achter naar voor (5 naar 1, alle cijfers 1-5)
		if digit == '5' && !hasUsedFive {
			newPrefix := string(digit) + prefix
			generateWithLastDigit(newPrefix, length-1, true, engines)
		} else if digit != '5' {
			newPrefix := string(digit) + prefix
			generateWithLastDigit(newPrefix, length-1, hasUsedFive, engines)
		}
	}
}

// generateEngines genereert alle engine codes met max 1 '5', van voor naar achter (geen restrictie op laatste cijfer)
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
		// Vul startDepth aan tot 12 cijfers met '5' als placeholder (genereren van achter naar voor, alle cijfers 1-5)
		prefix := ""
		for i := 0; i < 12-len(startDepth); i++ {
			prefix = "5" + prefix // Start met 5, genereren van achter naar voor
		}
		prefix = startDepth + prefix
		hasFive = strings.Contains(prefix, "5")
		// Genereer engines, van voor naar achter
		generateWithLastDigit(prefix, remainingLength, hasFive, &engines)
	} else {
		// Genereer alle engines van 12 posities, beginnend met 1-5, van voor naar achter
		for firstDigit := '5'; firstDigit >= '1'; firstDigit-- { // Genereren van achter naar voor (5 naar 1)
			prefix := string(firstDigit)
			hasFiveLocal := firstDigit == '5'
			// Genereer engines
			generateWithLastDigit(prefix, 11, hasFiveLocal, &engines)
		}
	}

	return engines
}

// simulateDepthGameToMoves genereert de zetten van een diepte-gebaseerde engine, reactief op de tegenstander
func simulateDepthGameToMoves(engine string, opponent string) (moves [13]byte) {
	if len(engine) != 12 || len(opponent) != 13 {
		return
	}

	p := Player{
		engineCode: engine,
		available:  [5]int{3, 3, 3, 3, 1},
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

// evaluateBatch evalueert een batch van engines met een expectedResult
func evaluateBatch(engines []string, inputEngines []string, expectedResult string, resultChan chan<- engineResult, progress *int32) {
	for _, engine := range engines {
		matches := 0
		totalScore := 0

		for _, inputEngine := range inputEngines {
			var result, p1Score, p2Score int
			if len(inputEngine) == 13 {
				if len(engine) == 12 {
					p1Moves := simulateDepthGameToMoves(engine, inputEngine)
					result, p1Score, p2Score = simulateFixedGame(string(p1Moves[:]), inputEngine)
				} else {
					result, p1Score, p2Score = simulateFixedGame(engine, inputEngine)
				}
			} else {
				result, p1Score, p2Score = simulateDepthGame(engine, inputEngine)
			}
			if result == -1 {
				continue // Skip invalid engines, but don't return early
			}
			if expectedResult == "Win" {
				if result == 1 {
					matches++
					totalScore += p1Score - p2Score
				}
			} else if expectedResult == "Draw" {
				if result == 0 {
					matches++
					totalScore += p1Score
				}
			} else if expectedResult == "Lose" {
				if result == 2 {
					matches++
					totalScore += p2Score - p1Score
				}
			}
		}

		if matches == len(inputEngines) {
			resultChan <- engineResult{engine: engine, score: totalScore}
		}
	}
	atomic.AddInt32(progress, int32(len(engines)))
}

// evaluateBatchClose evalueert een batch van engines voor "nooit verliezen"
func evaluateBatchClose(engines []string, inputEngines []string, resultChan chan<- engineResult, progress *int32) {
	for _, engine := range engines {
		neverLoses := true
		totalScore := 0

		for _, inputEngine := range inputEngines {
			var result, p1Score, p2Score int
			if len(inputEngine) == 13 {
				if len(engine) == 12 {
					p1Moves := simulateDepthGameToMoves(engine, inputEngine)
					result, p1Score, p2Score = simulateFixedGame(string(p1Moves[:]), inputEngine)
				} else {
					result, p1Score, p2Score = simulateFixedGame(engine, inputEngine)
				}
			} else {
				result, p1Score, p2Score = simulateDepthGame(engine, inputEngine)
			}
			if result == -1 || result == 2 { // Verlies of ongeldig
				neverLoses = false
				break
			}
			if result == 1 {
				totalScore += p1Score - p2Score
			} else { // Draw (result == 0)
				totalScore += p1Score
			}
		}

		if neverLoses {
			resultChan <- engineResult{engine: engine, score: totalScore}
		}
	}
	atomic.AddInt32(progress, int32(len(engines)))
}

// parseEngineCode haalt de engine code uit een invoer met prefix
func parseEngineCode(input string) string {
	parts := strings.Split(input, ":")
	if len(parts) > 2 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(input)
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

		var startDepth, expectedResult string
		var maxWorkers, maxMemoryMB int

		fmt.Println("Voer de startdepth in (leeg voor alle combinaties, bijv. '51'): ")
		fmt.Scanln(&startDepth)

		fmt.Println("Voer het verwachte resultaat in (Win/Draw/Lose): ")
		fmt.Scanln(&expectedResult)

		if len(startDepth) > 12 || (startDepth != "" && strings.ContainsAny(startDepth, "67890")) {
			fmt.Println("Ongeldige startdepth. Moet <= 12 chiffres zijn, alleen 1-5 of leeg.")
			continue
		}
		if expectedResult != "Win" && expectedResult != "Draw" && expectedResult != "Lose" {
			fmt.Println("Ongeldig resultaat. Gebruik 'Win', 'Draw' of 'Lose'.")
			continue
		}

		// Standaard 2x CPU-cores als aantal threads
		defaultWorkers := runtime.NumCPU() * 2
		fmt.Printf("Gebruik automatisch %d threads (2x aantal CPU-cores)? (ja/nee): ", defaultWorkers)
		var autoThreads string
		fmt.Scanln(&autoThreads)
		if strings.ToLower(autoThreads) == "ja" {
			maxWorkers = defaultWorkers
		} else {
			fmt.Println("Voer het aantal threads in (1-1000): ")
			var workersInput string
			fmt.Scanln(&workersInput)
			if n, err := fmt.Sscanf(workersInput, "%d", &maxWorkers); err != nil || n != 1 || maxWorkers < 1 || maxWorkers > 1000 {
				maxWorkers = defaultWorkers
				fmt.Printf("Ongeldige invoer, defaulting naar %d threads (2x aantal CPU-cores).\n", maxWorkers)
			}
		}

		fmt.Println("Voer het maximale geheugen in MB in (1-128000, default 100): ")
		var memoryInput string
		fmt.Scanln(&memoryInput)
		if n, err := fmt.Sscanf(memoryInput, "%d", &maxMemoryMB); err != nil || n != 1 || maxMemoryMB < 1 || maxMemoryMB > 128000 {
			maxMemoryMB = 100
			fmt.Println("Ongeldige invoer, defaulting naar 100 MB.")
		}

		generatedEngines := generateEngines(startDepth)

		const bytesPerResult = 24
		maxBufferSize := (maxMemoryMB * 1024 * 1024) / bytesPerResult // Bijv. 128,000 MB → ~5,368,709,120
		if maxBufferSize > len(generatedEngines) {
			maxBufferSize = len(generatedEngines)
		}
		if maxBufferSize < 10000 { // Verhoog minimum naar 10,000 voor grotere batches
			maxBufferSize = 10000
		}

		// Pre-allocate matching results
		matchingEngines := make([]engineResult, 0, 67108864) // Groter pre-alloceren voor alle mogelijke engines (max 1 '5')
		resultChan := make(chan engineResult, maxBufferSize)

		// Eerste poging: zoek engines die alles winnen
		var wg sync.WaitGroup
		var progress int32
		doneFirst := make(chan struct{})

		var startTime time.Time
		go func(totalEngines int) {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case t := <-ticker.C:
					p := atomic.LoadInt32(&progress)
					if t.Sub(startTime) >= 10*time.Second {
						speed := float64(p) / t.Sub(startTime).Seconds()
						fmt.Printf("Progress: %d / %d engines (%.2f%%), Speed: %.0f engines/s\n", p, totalEngines, float64(p)/float64(totalEngines)*100, speed)
					} else {
						fmt.Printf("Progress: %d / %d engines (%.2f%%)\n", p, totalEngines, float64(p)/float64(totalEngines)*100)
					}
				case <-doneFirst:
					p := atomic.LoadInt32(&progress)
					speed := float64(p) / time.Since(startTime).Seconds()
					fmt.Printf("Progress: %d / %d engines (%.2f%%), Speed: %.0f engines/s\n", p, totalEngines, float64(p)/float64(totalEngines)*100, speed)
					return
				}
			}
		}(len(generatedEngines))

		startTime = time.Now()
		batchSize := 1000
		for i := 0; i < len(generatedEngines); i += batchSize {
			end := i + batchSize
			if end > len(generatedEngines) {
				end = len(generatedEngines)
			}
			batch := generatedEngines[i:end]

			wg.Add(1)
			go func(engines []string) {
				defer wg.Done()
				evaluateBatch(engines, inputEngines, expectedResult, resultChan, &progress)
			}(batch)
		}

		go func() {
			wg.Wait()
			close(resultChan)
		}()

		for result := range resultChan {
			matchingEngines = append(matchingEngines, result)
		}
		close(doneFirst)

		if len(matchingEngines) > 0 {
			file, err := os.OpenFile("matching_engines.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Fout bij het openen van bestand: %v\n", err)
				continue
			}
			defer file.Close()

			sort.Slice(matchingEngines, func(i, j int) bool {
				return matchingEngines[i].score > matchingEngines[j].score
			})

			for _, result := range matchingEngines {
				_, err := file.WriteString(fmt.Sprintf("%s (score: %d)\n", result.engine, result.score))
				if err != nil {
					fmt.Printf("Fout bij het schrijven naar bestand: %v\n", err)
					break
				}
			}

			fmt.Printf("We found %d engines / %d total generated engines who %s from all input engines\n",
				len(matchingEngines), len(generatedEngines), strings.ToLower(expectedResult))
		} else {
			fmt.Printf("No engines found that %s against all input engines. Do you want to search for the best engine that never loses (win or draw)? (y/.): ", strings.ToLower(expectedResult))
			var response string
			fmt.Scanln(&response)
			if response == "y" {
				// Reset progress for the second run
				atomic.StoreInt32(&progress, 0)
				// Pre-allocate for close results
				matchingEnginesClose := make([]engineResult, 0, 67108864)
				resultChanClose := make(chan engineResult, maxBufferSize)

				var wgClose sync.WaitGroup
				doneClose := make(chan struct{})

				startTime = time.Now()
				go func(totalEngines int) {
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case t := <-ticker.C:
							p := atomic.LoadInt32(&progress)
							if t.Sub(startTime) >= 10*time.Second {
								speed := float64(p) / t.Sub(startTime).Seconds()
								fmt.Printf("Progress: %d / %d engines (%.2f%%), Speed: %.0f engines/s\n", p, totalEngines, float64(p)/float64(totalEngines)*100, speed)
							} else {
								fmt.Printf("Progress: %d / %d engines (%.2f%%)\n", p, totalEngines, float64(p)/float64(totalEngines)*100)
							}
						case <-doneClose:
							p := atomic.LoadInt32(&progress)
							speed := float64(p) / time.Since(startTime).Seconds()
							fmt.Printf("Progress: %d / %d engines (%.2f%%), Speed: %.0f engines/s\n", p, totalEngines, float64(p)/float64(totalEngines)*100, speed)
							return
						}
					}
				}(len(generatedEngines))

				for i := 0; i < len(generatedEngines); i += batchSize {
					end := i + batchSize
					if end > len(generatedEngines) {
						end = len(generatedEngines)
					}
					batch := generatedEngines[i:end]

					wgClose.Add(1)
					go func(engines []string) {
						defer wgClose.Done()
						evaluateBatchClose(engines, inputEngines, resultChanClose, &progress)
					}(batch)
				}

				go func() {
					wgClose.Wait()
					close(resultChanClose)
				}()

				for result := range resultChanClose {
					matchingEnginesClose = append(matchingEnginesClose, result)
				}
				close(doneClose)

				if len(matchingEnginesClose) > 0 {
					file, err := os.OpenFile("matching_engines.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Printf("Fout bij het openen van bestand: %v\n", err)
						continue
					}
					defer file.Close()

					sort.Slice(matchingEnginesClose, func(i, j int) bool {
						return matchingEnginesClose[i].score > matchingEnginesClose[j].score
					})

					for _, result := range matchingEnginesClose {
						_, err := file.WriteString(fmt.Sprintf("%s (score: %d)\n", result.engine, result.score))
						if err != nil {
							fmt.Printf("Fout bij het schrijven naar bestand: %v\n", err)
							break
						}
					}

					fmt.Printf("We found %d engines / %d total generated engines that never lose (win or draw) from all input engines\n",
						len(matchingEnginesClose), len(generatedEngines))
				} else {
					fmt.Println("No engines found that never lose against all input engines")
				}
			} else {
				fmt.Printf("No engines found that %s against all input engines\n", strings.ToLower(expectedResult))
			}
		}
	}
	fmt.Println("Gestopt.")
}
