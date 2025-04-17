package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	// Prompt templates
	promptAltCount         = "Введіть кількість альтернатив: "
	promptAltName          = "Введіть назву альтернативи %d: "
	promptStateCount       = "Введіть кількість зовнішніх умов (станів): "
	promptStateValue       = "Введіть значення корисності для альтернативи '%s' при стані %d (від 1 до %d): "
	promptMaxScore         = "Введіть максимальне значення бальної системи (наприклад, 10): "
	promptCriterionResults = "\nРезультати за критерієм %s:\n"

	// Error messages
	errInvalidCount = "Некоректне число %s"
	errInvalidScore = "Некоректне значення системи балів"
	errInvalidValue = "Некоректне значення. Будь ласка, спробуйте ще раз."

	// Table formats
	headerFormat      = "%-20s"
	stateHeaderFormat = "%-15s"
	scoreFormat       = "%-15.2f"
	resultRankFormat  = "%-5s %-20s %-15s\n"
	resultItemFormat  = "%-5d %-20s %-15.4f\n"
)

type (
	inputReader struct {
		reader *bufio.Reader
	}

	UncertainDecisionSystem struct {
		alternatives []string
		statesCount  int
		maxScore     int
		outcomes     map[string][]float64
	}

	// AltValue використовується для сортування альтернатив
	// по обчисленій величині критерію
	AltValue struct {
		alt   string
		value float64
	}
)

func newInputReader() *inputReader {
	return &inputReader{bufio.NewReader(os.Stdin)}
}

func (ir *inputReader) readString(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := ir.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (ir *inputReader) readInt(prompt string) (int, error) {
	str, err := ir.readString(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (ir *inputReader) readValidatedFloat(prompt string, min, max float64) float64 {
	for {
		str, err := ir.readString(prompt)
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(str, 64)
		if err == nil && val >= min && val <= max {
			return val
		}
		fmt.Println(errInvalidValue)
	}
}

func newUncertainDecisionSystem(ir *inputReader) (*UncertainDecisionSystem, error) {
	altCount, err := ir.readInt(promptAltCount)
	if err != nil || altCount <= 0 {
		return nil, fmt.Errorf(errInvalidCount, "альтернатив")
	}

	alts := make([]string, altCount)
	for i := range altCount {
		name, _ := ir.readString(fmt.Sprintf(promptAltName, i+1))
		alts[i] = name
	}

	stCount, err := ir.readInt(promptStateCount)
	if err != nil || stCount <= 0 {
		return nil, fmt.Errorf(errInvalidCount, "зовнішніх умов")
	}

	maxScore, err := ir.readInt(promptMaxScore)
	if err != nil || maxScore <= 0 {
		return nil, fmt.Errorf(errInvalidScore)
	}

	return &UncertainDecisionSystem{
		alternatives: alts,
		statesCount:  stCount,
		maxScore:     maxScore,
		outcomes:     make(map[string][]float64),
	}, nil
}

func (u *UncertainDecisionSystem) CollectOutcomes(ir *inputReader) {
	for _, alt := range u.alternatives {
		fmt.Printf("\nВведіть значення корисності для альтернативи '%s':\n", alt)
		values := make([]float64, u.statesCount)

		for j := range u.statesCount {
			prompt := fmt.Sprintf(promptStateValue, alt, j+1, u.maxScore)
			values[j] = ir.readValidatedFloat(prompt, 1, float64(u.maxScore))
		}

		u.outcomes[alt] = values
	}
}

func (u *UncertainDecisionSystem) PrintOutcomesMatrix() {
	fmt.Println("\nМатриця корисності:")
	fmt.Printf(headerFormat, "Альтернатива")
	for j := range u.statesCount {
		fmt.Printf(stateHeaderFormat, fmt.Sprintf("Стан %d", j+1))
	}
	fmt.Println()

	for _, alt := range u.alternatives {
		fmt.Printf(headerFormat, alt)
		for _, outcome := range u.outcomes[alt] {
			fmt.Printf(scoreFormat, outcome)
		}
		fmt.Println()
	}
}

// CalculateSavage розраховує критерій Севіджа:
// Для кожного стану знаходиться максимальне значення, після чого обчислюється "жалю"
// як різниця між максимальним значенням і значенням для альтернативи.
// Для кожної альтернативи береться максимальне значення жалю (мінімакс).
func (u *UncertainDecisionSystem) CalculateSavage() map[string]float64 {
	maxOutcomes := make([]float64, u.statesCount)

	// Знаходимо максимальне значення для кожного стану
	for j := range u.statesCount {
		maxVal := 0.0
		for _, alt := range u.alternatives {
			val := u.outcomes[alt][j]
			if val > maxVal {
				maxVal = val
			}
		}
		maxOutcomes[j] = maxVal
	}

	// Обчислюємо жалю для кожної альтернативи та знаходимо максимальне (найгірше)
	savage := make(map[string]float64)
	for _, alt := range u.alternatives {
		maxRegret := 0.0
		for j, outcome := range u.outcomes[alt] {
			regret := maxOutcomes[j] - outcome
			if regret > maxRegret {
				maxRegret = regret
			}
		}
		savage[alt] = maxRegret
	}
	return savage
}

// CalculateLaplace розраховує критерій Лапласа для кожної альтернативи
// як середнє значення по всіх станах (припускаючи, що всі стани рівноймовірні)
func (u *UncertainDecisionSystem) CalculateLaplace() map[string]float64 {
	laplace := make(map[string]float64)
	for _, alt := range u.alternatives {
		sum := 0.0
		for _, outcome := range u.outcomes[alt] {
			sum += outcome
		}

		avg := sum / float64(u.statesCount)
		laplace[alt] = avg
	}
	return laplace
}

func sortAltValues(data map[string]float64, ascending bool) []AltValue {
	arr := make([]AltValue, 0, len(data))
	for alt, val := range data {
		arr = append(arr, AltValue{alt, val})
	}
	// Для Севіджа (жалю) менше значення – краще; для Лапласа – більше значення – краще.
	if ascending {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].value < arr[j].value
		})
	} else {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].value > arr[j].value
		})
	}
	return arr
}

func PrintRanking(title string, altValues []AltValue, valueLabel string) {
	fmt.Printf(promptCriterionResults, title)
	fmt.Printf(resultRankFormat, "Ранг", "Альтернатива", valueLabel)
	for i, item := range altValues {
		fmt.Printf(resultItemFormat, i+1, item.alt, item.value)
	}
}

func main() {
	ir := newInputReader()
	u, err := newUncertainDecisionSystem(ir)
	if err != nil {
		fmt.Println(err)
		return
	}

	u.CollectOutcomes(ir)
	u.PrintOutcomesMatrix()

	// Розрахунок критерію Севіджа (мінімізація максимальної жалю)
	savage := u.CalculateSavage()
	sortedSev := sortAltValues(savage, true) // Нижче значення жалю – краще
	PrintRanking("Севіджа", sortedSev, "Макс. жалю")

	// Розрахунок критерію Лапласа (середнє значення корисності)
	laplace := u.CalculateLaplace()
	sortedLaplace := sortAltValues(laplace, false) // Вище середнє значення – краще
	PrintRanking("Лапласа", sortedLaplace, "Середня корисність")
}
