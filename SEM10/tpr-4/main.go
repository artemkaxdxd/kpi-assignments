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
	promptAltCount    = "Введіть кількість альтернатив: "
	promptAltName     = "Введіть назву альтернативи %d: "
	promptExpertCount = "Введіть кількість експертів: "
	promptExpertName  = "Введіть ім'я експерта %d: "
	promptRank        = "Ранг для альтернативи '%s' від експерта '%s' (1…%d): "

	colAltFormat    = "%-15s"
	colExpertFormat = "%-8s"
	colRankFormat   = "%-8d"
)

type (
	inputReader struct {
		r *bufio.Reader
	}

	ParetoSystem struct {
		alts      []string
		experts   []string
		rankings  map[string]map[string]int  // rankings[expert][alt] = rank
		dominance map[string]map[string]bool // dominance[a][b] = true якщо a домінує над b
	}
)

func newInputReader() *inputReader {
	return &inputReader{r: bufio.NewReader(os.Stdin)}
}

func (ir *inputReader) readString(prompt string) string {
	fmt.Print(prompt)
	s, _ := ir.r.ReadString('\n')
	return strings.TrimSpace(s)
}

func (ir *inputReader) readInt(prompt string) int {
	for {
		s := ir.readString(prompt)
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return v
		}
		fmt.Println("Невірне число, спробуйте ще раз.")
	}
}

func (ir *inputReader) readRank(prompt string, max int) int {
	for {
		s := ir.readString(prompt)
		if v, err := strconv.Atoi(s); err == nil && v >= 1 && v <= max {
			return v
		}
		fmt.Printf("Ведіть число від 1 до %d.\n", max)
	}
}

func newParetoSystem(ir *inputReader) *ParetoSystem {
	// Зчитуємо альтернативи
	n := ir.readInt(promptAltCount)
	alts := make([]string, n)
	for i := range n {
		alts[i] = ir.readString(fmt.Sprintf(promptAltName, i+1))
	}

	// Зчитуємо експертів
	n = ir.readInt(promptExpertCount)
	experts := make([]string, n)
	for i := range n {
		experts[i] = ir.readString(fmt.Sprintf(promptExpertName, i+1))
	}

	return &ParetoSystem{
		alts:      alts,
		experts:   experts,
		rankings:  make(map[string]map[string]int),
		dominance: make(map[string]map[string]bool),
	}
}

func (p *ParetoSystem) CollectRankings(ir *inputReader) {
	count := len(p.alts)

	for _, e := range p.experts {
		p.rankings[e] = make(map[string]int)
		fmt.Printf("\n--- Ранжування від експерта %s ---\n", e)

		for _, a := range p.alts {
			p.rankings[e][a] = ir.readRank(
				fmt.Sprintf(promptRank, a, e, count), count)
		}
	}
}

func (p *ParetoSystem) PrintRankingTable() {
	fmt.Println("\nТаблиця ранжувань (рядок – альтернатива, стовпці – експерти):")

	fmt.Printf(colAltFormat, "Альтернатива")
	for _, e := range p.experts {
		fmt.Printf(colExpertFormat, e)
	}
	fmt.Println()

	for _, a := range p.alts {
		fmt.Printf(colAltFormat, a)
		for _, e := range p.experts {
			fmt.Printf(colRankFormat, p.rankings[e][a])
		}
		fmt.Println()
	}
}

func (p *ParetoSystem) BuildDominance() {
	for _, a := range p.alts {
		p.dominance[a] = make(map[string]bool)
	}

	for _, a1 := range p.alts {
		for _, a2 := range p.alts {
			if a1 == a2 {
				continue
			}

			better := false
			notWorse := true

			for _, e := range p.experts {
				r1 := p.rankings[e][a1]
				r2 := p.rankings[e][a2]

				if r1 > r2 {
					notWorse = false
					break
				}

				if r1 < r2 {
					better = true
				}
			}

			if notWorse && better {
				p.dominance[a1][a2] = true
			}
		}
	}
}

func (p *ParetoSystem) PrintDominanceMatrix() {
	fmt.Println("\nМатриця домінування (1 – рядок домінує над стовпцем):")

	fmt.Printf(colAltFormat, "")
	for _, a := range p.alts {
		fmt.Printf("%-8s", a)
	}
	fmt.Println()

	for _, a1 := range p.alts {
		fmt.Printf(colAltFormat, a1)
		for _, a2 := range p.alts {
			if a1 == a2 {
				fmt.Printf("%-8s", "-")
			} else if p.dominance[a1][a2] {
				fmt.Printf("%-8d", 1)
			} else {
				fmt.Printf("%-8d", 0)
			}
		}
		fmt.Println()
	}
}

func (p *ParetoSystem) ParetoSet() []string {
	out := []string{}
	for _, a := range p.alts {
		dominated := false

		for _, b := range p.alts {
			if p.dominance[b][a] {
				dominated = true
				break
			}
		}

		if !dominated {
			out = append(out, a)
		}
	}

	sort.Strings(out)
	return out
}

func main() {
	ir := newInputReader()
	ps := newParetoSystem(ir)

	ps.CollectRankings(ir)
	ps.PrintRankingTable()

	ps.BuildDominance()
	ps.PrintDominanceMatrix()

	pareto := ps.ParetoSet()
	fmt.Println("\nМножина Парето оптимальних альтернатив:")
	for i, a := range pareto {
		fmt.Printf("%d) %s\n", i+1, a)
	}
}
