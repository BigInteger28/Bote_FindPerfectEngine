package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// elementsDepth definieert de transformaties van elementen op verschillende dieptes
var elementsDepth = map[byte]map[int]byte{
	'W': {1: 'L', 2: 'A', 3: 'V', 4: 'W'},
	'V': {1: 'W', 2: 'L', 3: 'A', 4: 'V'},
	'A': {1: 'V', 2: 'W', 3: 'L', 4: 'A'},
	'L': {1: 'A', 2: 'V', 3: 'W', 4: 'L'},
}

// Player houdt de staat van een speler bij
type Player struct {
	engineCode string
	available  map[byte]int
	moves      []byte
}

// getElementFromCode haalt direct een element op basis van de engine code voor de eerste zet
func getElementFromCode(depth int) byte {
	switch depth {
	case 1:
		return 'W'
	case 2:
		return 'V'
	case 3:
		return 'A'
	case 4:
		return 'L'
	case 5:
		return 'D'
	default:
		return 'W'
	}
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
func chooseAvailableElement(target byte, available map[byte]int) byte {
	if target == 'D' && available['D'] > 0 {
		return 'D'
	}
	if target != 0 && available[target] > 0 {
		return target
	}
	current := target
	for i := 0; i < 5; i++ {
		current = elementsDepth[current][1]
		if available[current] > 0 {
			return current
		}
	}
	if available['D'] > 0 {
		return 'D'
	}
	return 0
}

// getLastElement bepaalt het resterende element voor de 13e zet
func getLastElement(available map[byte]int) byte {
	for elem, count := range available {
		if count > 0 {
			return elem
		}
	}
	return 0
}

// determineWinner bepaalt de winnaar van een zet
func determineWinner(move1, move2 byte) int {
	if move1 == move2 {
		return 0
	}
	switch move1 {
	case 'W':
		if move2 == 'V' {
			return 1
		}
		if move2 == 'L' {
			return 2
		}
	case 'V':
		if move2 == 'A' {
			return 1
		}
		if move2 == 'W' {
			return 2
		}
	case 'A':
		if move2 == 'L' {
			return 1
		}
		if move2 == 'V' {
			return 2
		}
	case 'L':
		if move2 == 'W' {
			return 1
		}
		if move2 == 'A' {
			return 2
		}
	}
	return 0
}

// simulateDepthGame simuleert een spel met diepte-gebaseerde codes
func simulateDepthGame(engine1, engine2 string) (result int, p1Score, p2Score int) {
	if len(engine1) != 12 || len(engine2) != 12 {
		return -1, 0, 0
	}

	p1 := Player{
		engineCode: engine1,
		available:  map[byte]int{'W': 3, 'V': 3, 'A': 3, 'L': 3, 'D': 1},
		moves:      make([]byte, 0, 13),
	}
	p2 := Player{
		engineCode: engine2,
		available:  map[byte]int{'W': 3, 'V': 3, 'A': 3, 'L': 3, 'D': 1},
		moves:      make([]byte, 0, 13),
	}

	p1Score, p2Score = 0, 0

	for i := 0; i < 12; i++ {
		depth1 := int(engine1[i] - '0')
		depth2 := int(engine2[i] - '0')

		var move1, move2 byte
		if i == 0 {
			move1 = chooseAvailableElement(getElementFromCode(depth1), p1.available)
			move2 = chooseAvailableElement(getElementFromCode(depth2), p2.available)
		} else {
			move1 = chooseAvailableElement(getElementByDepth(p2.moves[i-1], depth1), p1.available)
			move2 = chooseAvailableElement(getElementByDepth(p1.moves[i-1], depth2), p2.available)
		}

		if move1 != 0 {
			p1.available[move1]--
			p1.moves = append(p1.moves, move1)
		}
		if move2 != 0 {
			p2.available[move2]--
			p2.moves = append(p2.moves, move2)
		}

		winner := determineWinner(move1, move2)
		if winner == 1 {
			p1Score++
		} else if winner == 2 {
			p2Score++
		}
	}

	move1 := getLastElement(p1.available)
	move2 := getLastElement(p2.available)
	p1.available[move1]--
	p2.available[move2]--
	p1.moves = append(p1.moves, move1)
	p2.moves = append(p2.moves, move2)

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
		move1 := engine1[i]
		move2 := engine2[i]

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

// generateEngines genereert alle engine codes met max 1 '5'
func generateEngines(startDepth string) []string {
	var engines []string
	remainingLength := 12 - len(startDepth)
	hasFive := strings.Contains(startDepth, "5")

	if remainingLength < 0 {
		return engines
	}

	if hasFive {
		totalCombinations := intPow(4, remainingLength)
		engines = make([]string, 0, totalCombinations)
		for i := 0; i < totalCombinations; i++ {
			suffix := base4ToDecimal(i, remainingLength)
			engine := startDepth + suffix
			engines = append(engines, engine)
		}
	} else {
		totalCombinations := intPow(4, remainingLength)
		extraCombinations := remainingLength * intPow(4, remainingLength-1)
		engines = make([]string, 0, totalCombinations+extraCombinations)
		for i := 0; i < totalCombinations; i++ {
			suffix := base4ToDecimal(i, remainingLength)
			engine := startDepth + suffix
			engines = append(engines, engine)
		}

		for pos := 0; pos < remainingLength; pos++ {
			baseCombinations := intPow(4, remainingLength-1)
			for i := 0; i < baseCombinations; i++ {
				suffix := base4ToDecimal(i, remainingLength-1)
				engine := startDepth + suffix[:pos] + "5" + suffix[pos:]
				engines = append(engines, engine)
			}
		}
	}

	if startDepth == "" {
		engines = []string{}
		totalCombinations := intPow(4, 12)
		extraCombinations := 12 * intPow(4, 11)
		engines = make([]string, 0, totalCombinations+extraCombinations)
		for i := 0; i < totalCombinations; i++ {
			engine := base4ToDecimal(i, 12)
			engines = append(engines, engine)
		}
		for pos := 0; pos < 12; pos++ {
			baseCombinations := intPow(4, 11)
			for i := 0; i < baseCombinations; i++ {
				suffix := base4ToDecimal(i, 11)
				engine := suffix[:pos] + "5" + suffix[pos:]
				engines = append(engines, engine)
			}
		}
	}

	return engines
}

// base4ToDecimal converteert een decimaal naar een base-4 string (1-4)
func base4ToDecimal(num, length int) string {
	digits := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		digits[i] = byte((num % 4) + 1 + '0')
		num /= 4
	}
	return string(digits)
}

// intPow berekent base^exp
func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// engineResult houdt een engine en zijn totaalscore bij
type engineResult struct {
	engine string
	score  int
}

// simulateDepthGameToMoves genereert de zetten van een diepte-gebaseerde engine, reactief op de tegenstander
func simulateDepthGameToMoves(engine string, opponent string) []byte {
	if len(engine) != 12 || len(opponent) != 13 {
		return nil
	}

	p := Player{
		engineCode: engine,
		available:  map[byte]int{'W': 3, 'V': 3, 'A': 3, 'L': 3, 'D': 1},
		moves:      make([]byte, 0, 13),
	}

	for i := 0; i < 12; i++ {
		depth := int(engine[i] - '0')
		var move byte
		if i == 0 {
			move = chooseAvailableElement(getElementFromCode(depth), p.available)
		} else {
			// Gebruik de vorige zet van de tegenstander
			move = chooseAvailableElement(getElementByDepth(opponent[i-1], depth), p.available)
		}
		if move != 0 {
			p.available[move]--
			p.moves = append(p.moves, move)
		} else {
			for elem, count := range p.available {
				if count > 0 {
					move = elem
					p.available[move]--
					p.moves = append(p.moves, move)
					break
				}
			}
		}
	}

	move := getLastElement(p.available)
	if move != 0 {
		p.available[move]--
		p.moves = append(p.moves, move)
	} else {
		p.moves = append(p.moves, 'W')
	}

	return p.moves
}

// evaluateEngine simuleert een engine en berekent de totaalscore (voor "Win" tegen alles)
func evaluateEngine(engine string, inputEngines []string, expectedResult string, resultChan chan<- engineResult, progress *int, mu *sync.Mutex) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
		mu.Lock()
		*progress++
		mu.Unlock()
	}()

	matches := 0
	totalScore := 0

	for _, inputEngine := range inputEngines {
		var result, p1Score, p2Score int
		if len(inputEngine) == 13 {
			if len(engine) == 12 {
				p1Moves := simulateDepthGameToMoves(engine, inputEngine)
				if len(p1Moves) == 13 {
					result, p1Score, p2Score = simulateFixedGame(string(p1Moves), inputEngine)
				} else {
					return
				}
			} else {
				result, p1Score, p2Score = simulateFixedGame(engine, inputEngine)
			}
		} else {
			result, p1Score, p2Score = simulateDepthGame(engine, inputEngine)
		}
		if result == -1 {
			return
		}
		switch expectedResult {
		case "Win":
			if result != 1 { // Vroegtijdig stoppen bij geen winst
				return
			}
			matches++
			totalScore += p1Score - p2Score
		case "Draw":
			if result != 0 {
				return
			}
			matches++
			totalScore += p1Score
		case "Lose":
			if result != 2 {
				return
			}
			matches++
			totalScore += p2Score - p1Score
		}
	}

	if matches == len(inputEngines) {
		resultChan <- engineResult{engine: engine, score: totalScore}
	}
}

// evaluateEngineClose simuleert een engine en berekent de totaalscore (nooit verliezen, winnen of draw)
func evaluateEngineClose(engine string, inputEngines []string, resultChan chan<- engineResult, progress *int, mu *sync.Mutex) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
		mu.Lock()
		*progress++
		mu.Unlock()
	}()

	totalScore := 0

	for _, inputEngine := range inputEngines {
		var result, p1Score, p2Score int
		if len(inputEngine) == 13 {
			if len(engine) == 12 {
				p1Moves := simulateDepthGameToMoves(engine, inputEngine)
				if len(p1Moves) != 13 {
					return
				}
				result, p1Score, p2Score = simulateFixedGame(string(p1Moves), inputEngine)
			} else {
				result, p1Score, p2Score = simulateFixedGame(engine, inputEngine)
			}
		} else {
			result, p1Score, p2Score = simulateDepthGame(engine, inputEngine)
		}
		if result == -1 {
			return
		}
		if result == 2 { // Verlies
			return // Stop immediately on loss
		}
		if result == 1 { // Win
			totalScore += p1Score - p2Score
		} else if result == 0 { // Draw
			totalScore += p1Score
		}
	}

	resultChan <- engineResult{engine: engine, score: totalScore}
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

		fmt.Println("Voer het aantal threads in (1-1000, default [CPU-cores]): ")
		var workersInput string
		fmt.Scanln(&workersInput)
		if n, err := fmt.Sscanf(workersInput, "%d", &maxWorkers); err != nil || n != 1 || maxWorkers < 1 || maxWorkers > 1000 {
			maxWorkers = runtime.NumCPU() // Default naar aantal CPU-cores
			fmt.Printf("Ongeldige invoer, defaulting naar %d threads (aantal CPU-cores).\n", maxWorkers)
		}

		fmt.Println("Voer het maximale geheugen in MB in (1-128000, default 100): ")
		var memoryInput string
		fmt.Scanln(&memoryInput)
		if n, err := fmt.Sscanf(memoryInput, "%d", &maxMemoryMB); err != nil || n != 1 || maxMemoryMB < 1 || maxMemoryMB > 128000 {
			maxMemoryMB = 100 // Default
			fmt.Println("Ongeldige invoer, defaulting naar 100 MB.")
		}

		generatedEngines := generateEngines(startDepth)

		// Bereken kanaalbuffer op basis van geheugen (elk engineResult is ~24 bytes)
		const bytesPerResult = 24
		maxBufferSize := (maxMemoryMB * 1024 * 1024) / bytesPerResult
		if maxBufferSize > len(generatedEngines) {
			maxBufferSize = len(generatedEngines) // Beperk tot aantal engines
		}
		if maxBufferSize < 1000 {
			maxBufferSize = 1000 // Minimum buffer
		}

		// Eerste poging: zoek engines die alles winnen
		resultChan := make(chan engineResult, maxBufferSize)
		var wg sync.WaitGroup
		var matchingEngines []engineResult
		workerChan := make(chan struct{}, maxWorkers)
		var progress int
		var mu sync.Mutex
		doneFirst := make(chan struct{})

		go func(totalEngines int) {
			ticker := time.NewTicker(5 * time.Second) // Elke 5 seconden update
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					mu.Lock()
					fmt.Printf("Progress: %d / %d engines (%.2f%%)\n", progress, totalEngines, float64(progress)/float64(totalEngines)*100)
					mu.Unlock()
				case <-doneFirst:
					return
				}
			}
		}(len(generatedEngines))

		for _, engine := range generatedEngines {
			wg.Add(1)
			workerChan <- struct{}{}
			go func(e string) {
				defer wg.Done()
				defer func() { <-workerChan }()
				evaluateEngine(e, inputEngines, expectedResult, resultChan, &progress, &mu)
			}(engine)
		}

		go func() {
			wg.Wait()
			close(resultChan)
		}()

		for result := range resultChan {
			matchingEngines = append(matchingEngines, result)
		}
		close(doneFirst)

		// Sorteer op totaalscore
		if expectedResult == "Win" || expectedResult == "Lose" {
			sort.Slice(matchingEngines, func(i, j int) bool {
				return matchingEngines[i].score > matchingEngines[j].score
			})
		} else if expectedResult == "Draw" {
			sort.Slice(matchingEngines, func(i, j int) bool {
				return matchingEngines[i].score > matchingEngines[j].score
			})
		}

		// Schrijf naar bestand en toon resultaat
		if len(matchingEngines) > 0 {
			file, err := os.OpenFile("matching_engines.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Fout bij het openen van bestand: %v\n", err)
				continue
			}
			defer file.Close()

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
			fmt.Printf("No wins to all found. Do you want to search for the best engine close to all wins? (y/.): ")
			var response string
			fmt.Scanln(&response)
			if response == "y" {
				// Tweede poging: zoek engines die nooit verliezen
				resultChanClose := make(chan engineResult, maxBufferSize)
				var wgClose sync.WaitGroup
				var matchingEnginesClose []engineResult
				progress = 0 // Reset progress
				doneClose := make(chan struct{})

				go func(totalEngines int) {
					ticker := time.NewTicker(5 * time.Second) // Elke 5 seconden update
					defer ticker.Stop()
					for {
						select {
						case <-ticker.C:
							mu.Lock()
							fmt.Printf("Progress: %d / %d engines (%.2f%%)\n", progress, totalEngines, float64(progress)/float64(totalEngines)*100)
							mu.Unlock()
						case <-doneClose:
							return
						}
					}
				}(len(generatedEngines))

				for _, engine := range generatedEngines {
					wgClose.Add(1)
					workerChan <- struct{}{}
					go func(e string) {
						defer wgClose.Done()
						defer func() { <-workerChan }()
						evaluateEngineClose(e, inputEngines, resultChanClose, &progress, &mu)
					}(engine)
				}

				go func() {
					wgClose.Wait()
					close(resultChanClose)
				}()

				for result := range resultChanClose {
					matchingEnginesClose = append(matchingEnginesClose, result)
				}
				close(doneClose)

				// Sorteer op totaalscore
				sort.Slice(matchingEnginesClose, func(i, j int) bool {
					return matchingEnginesClose[i].score > matchingEnginesClose[j].score
				})

				if len(matchingEnginesClose) > 0 {
					file, err := os.OpenFile("matching_engines.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Printf("Fout bij het openen van bestand: %v\n", err)
						continue
					}
					defer file.Close()

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
