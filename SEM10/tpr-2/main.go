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
	promptAltCount         = "Введіть кількість альтернатив: "
	promptAltName          = "Введіть назву альтернативи %d: "
	promptAltValue         = "\nВведіть значення корисності для альтернативи '%s':\n"
	promptStateCount       = "Введіть кількість зовнішніх умов (станів): "
	promptStateValue       = "Введіть значення корисності для альтернативи '%s' при стані %d (від 1 до %d): "
	promptMaxScore         = "Введіть максимальне значення бальної системи (наприклад, 10): "
	promptAlpha            = "Введіть коефіцієнт оптимізму α (від 0 до 1): "
	promptCriterionResults = "\nРезультати за критерієм %s:\n"

	errInvalidCount = "Некоректне число %s"
	errInvalidScore = "Некоректне значення системи балів"
	errInvalidValue = "Некоректне значення. Будь ласка, спробуйте ще раз."

	headerFormat      = "%-20s"
	altHeaderFormat   = "%-20s"
	stateHeaderFormat = "%-15s"
	scoreFormat       = "%-15.2f"
	resultRankFormat  = "%-5s %-20s %-15s\n"
	resultItemFormat  = "%-5d %-20s %-15.4f\n"
)

type (
	inputReader struct {
		reader *bufio.Reader
	}

	Alternative struct {
		name    string
		wald    float64 // мінімальне значення
		maxmax  float64 // максимальне значення
		hurwicz float64 // критерій Гурвіца
	}

	UncertainDecisionSystem struct {
		alternatives []string
		statesCount  int
		maxScore     int
		// outcomes maps alternative name to slice of outcomes
		outcomes map[string][]float64
	}

	ByCriterion struct {
		alts  []Alternative
		value func(a Alternative) float64
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
	input, err := ir.readString(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(input)
}

func (ir *inputReader) readFloat(prompt string) (float64, error) {
	input, err := ir.readString(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(input, 64)
}

func (ir *inputReader) readStringArray(count int, promptTemplate string) []string {
	items := make([]string, count)
	for i := range count {
		prompt := fmt.Sprintf(promptTemplate, i+1)
		str, _ := ir.readString(prompt)
		items[i] = str
	}
	return items
}

func (ir *inputReader) readValidatedFloat(prompt string, min, max float64) float64 {
	for {
		value, err := ir.readFloat(prompt)
		if err == nil && value >= min && value <= max {
			return value
		}
		fmt.Println(errInvalidValue)
	}
}

func newUncertainDecisionSystem(ir *inputReader) (*UncertainDecisionSystem, error) {
	altCount, err := ir.readInt(promptAltCount)
	if err != nil || altCount <= 0 {
		return nil, fmt.Errorf(errInvalidCount, "альтернатив")
	}

	alternatives := ir.readStringArray(altCount, promptAltName)

	stateCount, err := ir.readInt(promptStateCount)
	if err != nil || stateCount <= 0 {
		return nil, fmt.Errorf(errInvalidCount, "зовнішніх умов")
	}

	maxScore, err := ir.readInt(promptMaxScore)
	if err != nil || maxScore <= 0 {
		return nil, fmt.Errorf(errInvalidScore)
	}

	return &UncertainDecisionSystem{
		alternatives: alternatives,
		statesCount:  stateCount,
		maxScore:     maxScore,
		outcomes:     make(map[string][]float64),
	}, nil
}

func (u *UncertainDecisionSystem) CollectOutcomes(ir *inputReader) {
	for _, alt := range u.alternatives {
		fmt.Printf(promptAltValue, alt)
		outcomeSlice := make([]float64, u.statesCount)

		for j := range u.statesCount {
			prompt := fmt.Sprintf(promptStateValue, alt, j+1, u.maxScore)
			outcomeSlice[j] = ir.readValidatedFloat(prompt, 1, float64(u.maxScore))
		}

		u.outcomes[alt] = outcomeSlice
	}
}

func (u *UncertainDecisionSystem) PrintOutcomesMatrix() {
	fmt.Println("\nМатриця корисності альтернатив для кожного стану:")
	fmt.Printf(headerFormat, "Альтернатива")

	for j := range u.statesCount {
		fmt.Printf(stateHeaderFormat, fmt.Sprintf("Стан %d", j+1))
	}
	fmt.Println()

	for _, alt := range u.alternatives {
		fmt.Printf(altHeaderFormat, alt)
		for _, outcome := range u.outcomes[alt] {
			fmt.Printf(scoreFormat, outcome)
		}
		fmt.Println()
	}
}

func (u *UncertainDecisionSystem) CalculateCriteria(ir *inputReader) []Alternative {
	alpha := ir.readValidatedFloat(promptAlpha, 0, 1)
	alts := make([]Alternative, len(u.alternatives))

	for i, alt := range u.alternatives {
		data := u.outcomes[alt]
		if len(data) == 0 {
			continue
		}

		minVal, maxVal := data[0], data[0]
		for _, v := range data {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}

		hurwicz := alpha*maxVal + (1-alpha)*minVal

		alts[i] = Alternative{
			name:    alt,
			wald:    minVal,
			maxmax:  maxVal,
			hurwicz: hurwicz,
		}
	}
	return alts
}

func (u *UncertainDecisionSystem) PrintRankings(criterionName string, alts []Alternative, valueFunc func(a Alternative) float64) {
	sort.Sort(ByCriterion{alts: alts, value: valueFunc})

	fmt.Printf(promptCriterionResults, criterionName)
	fmt.Printf(resultRankFormat, "Ранг", "Альтернатива", criterionName)

	for i, alt := range alts {
		fmt.Printf(resultItemFormat, i+1, alt.name, valueFunc(alt))
	}
}

func (b ByCriterion) Len() int           { return len(b.alts) }
func (b ByCriterion) Swap(i, j int)      { b.alts[i], b.alts[j] = b.alts[j], b.alts[i] }
func (b ByCriterion) Less(i, j int) bool { return b.value(b.alts[i]) > b.value(b.alts[j]) }

func main() {
	ir := newInputReader()
	u, err := newUncertainDecisionSystem(ir)
	if err != nil {
		fmt.Println(err)
		return
	}

	u.CollectOutcomes(ir)
	u.PrintOutcomesMatrix()

	alts := u.CalculateCriteria(ir)

	u.PrintRankings("Вальда", alts, func(a Alternative) float64 { return a.wald })
	u.PrintRankings("maxmax", alts, func(a Alternative) float64 { return a.maxmax })
	u.PrintRankings("Гурвіца", alts, func(a Alternative) float64 { return a.hurwicz })
}
