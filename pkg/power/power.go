package power

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sajari/regression"
	"gonum.org/v1/gonum/floats"
)

func NewFormula() *FormulaProvider {
	fp := &FormulaProvider{
		Formula:      formula{},
		HasFormula:   false,
		FormulaSlice: make([][]string, 0),
		PowerChan:    make([]float64, 0),
	}

	return fp
}

func (fp *FormulaProvider) Regression(start [][]string) (a float64, b float64, intercept float64) {
	records := make([][]float64, 0, len(start))
	powerslice := make([]float64, 0, len(start))
	cpuslice := make([]float64, 0, len(start))
	memslice := make([]float64, 0, len(start))
	for i, s := range start {
		r := make([]float64, 0, len(s))
		for _, atom := range s {
			record, err := strconv.ParseFloat(atom, 64)
			if err != nil {
				log.Println(err)
			}
			r = append(r, record)
		}
		fmt.Println(r)
		records = append(records, r)
		powerslice = append(powerslice, records[i][0])
		cpuslice = append(cpuslice, records[i][1])
		memslice = append(memslice, records[i][2])
	}
	powerMax := floats.Max(powerslice)
	powerMin := floats.Min(powerslice)
	cpuMax := floats.Max(cpuslice)
	cpuMin := floats.Min(cpuslice)

	memMax := floats.Max(memslice)
	memMin := floats.Min(memslice)

	fmt.Println(powerMax)
	fmt.Println(memMin)
	log.Println("recieve records -> len =", len(records))

	fp.Formula.Regression.SetObserved("Power")
	fp.Formula.Regression.SetVar(0, "Cpu")
	fp.Formula.Regression.SetVar(1, "Memory")

	for i, record := range records {
		power := records[i][0]
		power = (power - powerMin) / (powerMax - powerMin)
		cpu := record[1]
		cpu = (cpu - cpuMin) / (cpuMax - cpuMin)
		memory := record[2]
		memory = (memory - memMin) / (memMax - memMin)

		fp.Formula.Regression.Train(regression.DataPoint(power, []float64{cpu, memory}))
	}
	// Train/fit the regression model.
	err := fp.Formula.Regression.Run()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("\nRegression Formula:\n%v\n\n", fp.Formula.Regression.Formula)
	err = fp.Formula.getCoefficient(fp.Formula.Regression.Formula)
	if err != nil {
		log.Println(err)
	}

	return a, b, intercept
}

func (f *formula) getCoefficient(formula string) (err error) {
	temp := strings.Split(formula, " = ")
	spstring := strings.Split(temp[1], " + ")
	f.Intercept, err = strconv.ParseFloat(spstring[0], 64)
	if err != nil {
		return err
	}
	spstring = spstring[1:]
	for _, atom := range spstring {
		temp := strings.Split(atom, "*")
		metric := temp[0]
		strValue := temp[1]
		switch metric {
		case "Cpu":
			f.Alpha, err = strconv.ParseFloat(strValue, 64)
			if err != nil {
				return err
			}
			log.Println(f.Alpha)

		case "Memory":
			f.Beta, err = strconv.ParseFloat(strValue, 64)
			if err != nil {
				return err
			}
			log.Println(f.Beta)
		}

	}

	return nil

}

func (fp *FormulaProvider) GetPower(powerchan chan float64) {

	cmd := exec.Command("turbostat", "--Summary", "-i", "1", "-n", "1", "-s", "PkgWatt")

	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	slice := strings.Split(string(out), "PkgWatt\n")

	var a string
	for _, str := range slice {
		a = str
	}

	var b string
	b = a[0:5]
	// secondslice := strings.Split(a, " ")

	s, err := strconv.ParseFloat(b, 64)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(s)

	powerchan <- s

}
