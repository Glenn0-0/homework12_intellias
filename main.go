package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"errors" // пакети для обробки помилок.

	"encoding/json" // пакети для роботи з JSON.
	"io/ioutil"
)

type Trains []Train

type Train struct {
	TrainID            int       `json:"trainId"`
	DepartureStationID int       `json:"departureStationId"`
	ArrivalStationID   int       `json:"arrivalStationId"`
	Price              float32   `json:"price"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	DepartureTime      time.Time `json:"departureTime"`
}

var (
	validCriteria = []string{"price", "departure-time", "arrival-time"} // список можливих критеріїв.

	criteriaArrTime = validCriteria[2]
	criteriaDepTime = validCriteria[1]
	criteriaPrice   = validCriteria[0]
)

const (
	dataFile           = "data.json" // назва файлу з рейсами.
	maxNumberOfResults = 3           // кількість рейсів, що потрібно вивести.
	timeLayout         = "15:04:05"  // макет для форматування часу.
)

func main() {
	//	... запит даних від користувача.
	var departureStation, arrivalStation, criteria string

	fmt.Println("Please, specify your departure info:")

	fmt.Print("departureStation: ")
	fmt.Scanln(&departureStation)

	fmt.Print("arrivalStation: ")
	fmt.Scanln(&arrivalStation)

	fmt.Printf(`!! -- Keep in mind: valid criteria values are "%s", "%s", "%s" (without quotes).%s`, criteriaPrice, criteriaDepTime, criteriaArrTime, "\n")
	fmt.Print("criteria: ")
	fmt.Scanln(&criteria)

	// знаходимо 3 перші (за сортуванням) рейси, що б відповідали заданим значенням depStation та arrStation.
	result, err := FindTrains(departureStation, arrivalStation, criteria)

	//	... обробка помилки.
	if err != nil {
		fmt.Println(err)
		return
	}

	//	... друк result.
	result.printTrains()
}

// повертає 3 перші за сортуванням поїзди, що задовільняють станції прибуття та відправлення; перевіряє валідність вхідних даних користувача.
func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	// обробка помилок з вхідних даних.
	// перевірка на непусті значення станцій та валідний критерій
	errInput := checkInput(departureStation, arrivalStation, criteria)
	if errInput != nil {
		return nil, errInput // мало б бути fmt.Errorf("invalid input: %w", errInput), але нехай вже буде виправлене.
	}

	// конвертування у int та перевірка на валідне значення станції.
	departureStationID, errDepStation := strconv.Atoi(departureStation)
	if departureStationID < 0 || errDepStation != nil { // перевірка на натуральне число.
		return nil, errors.New("bad departure station input")
	}

	arrivalStationID, errArrStation := strconv.Atoi(arrivalStation)
	if arrivalStationID < 0 || errArrStation != nil { // перевірка на натуральне число.
		return nil, errors.New("bad arrival station input")
	}

	// ... код
	// починаємо парсити файл JSON та обробляємо помилки, якщо виникають.
	trains, errParse := parseJSON()
	if errParse != nil {
		return nil, errParse
	}

	// знаходимо усі рейси, що б відповідали заданим значенням depStation та arrStation та сортуємо варіанти за критерієм.
	results := sortByCriteria(getTrains(trains, departureStationID, arrivalStationID), criteria)

	// маєте повернути перші 3 правильні значення (або ніл, якщо їх нема).
	if len(results) < maxNumberOfResults {
		return results, nil
	}

	return results[:maxNumberOfResults], nil
}

// перевіряє валідність вхідних даних, повертає числові значення станцій або помилку.
func checkInput(depStation, arrStation, criteria string) error {
	// перевірка на непусте значення станції відправлення.
	if depStation == "" {
		return errors.New("empty departure station")
	}

	// перевірка на непусте значення станції прибуття.
	if arrStation == "" {
		return errors.New("empty arrival station")
	}

	// перевірка на валідний критерій для сортування.
	if !contains(validCriteria, criteria) {
		return errors.New("unsupported criteria")
	}

	return nil
}

func contains(sliceOfCriterias []string, criteria string) bool {
	for _, cr := range sliceOfCriterias {
		if criteria == cr {
			return true
		}
	}

	return false
}

// метод для анмаршалінгу JSON, адже "hh:mm:ss" не парситься у поле типу time.Time.
func (tr *Train) UnmarshalJSON(data []byte) error {
	type TrainStringTime struct {
		TrainID            int     `json:"trainId"`
		DepartureStationID int     `json:"departureStationId"`
		ArrivalStationID   int     `json:"arrivalStationId"`
		Price              float32 `json:"price"`
		ArrivalTime        string  `json:"arrivalTime"`
		DepartureTime      string  `json:"departureTime"`
	}

	// анмаршал у тимчасову структуру, зі зручним типом.
	var res TrainStringTime
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	// парсинг часу відправлення та прибуття; обробка помилок.
	parsedDepTime, errDepTime := time.Parse(timeLayout, res.DepartureTime)
	if errDepTime != nil {
		return fmt.Errorf("failed to parse departure time: %w", errDepTime)
	}

	parsedArrTime, errArrTime := time.Parse(timeLayout, res.ArrivalTime)
	if errArrTime != nil {
		return fmt.Errorf("failed to parse arrival time: %w", errArrTime)
	}

	// присвоюємо потрібній структурі значення з тимчасової, конвертуючи.
	tr.DepartureTime = time.Date(0, time.January, 1, parsedDepTime.Hour(), parsedDepTime.Minute(), parsedDepTime.Second(), 0, time.UTC)
	tr.ArrivalTime = time.Date(0, time.January, 1, parsedArrTime.Hour(), parsedArrTime.Minute(), parsedArrTime.Second(), 0, time.UTC)

	// присвоюємо усі інші значення (без змін).
	tr.TrainID = res.TrainID
	tr.DepartureStationID = res.DepartureStationID
	tr.ArrivalStationID = res.ArrivalStationID
	tr.Price = res.Price

	return nil // повернеться за відсутності помилок.
}

// анмаршал JSON файла у Trains.
func parseJSON() ([]Train, error) {
	// відкриваємо та читаємо файл, якщо не вийшло - повертаємо помилку.
	data, errRead := ioutil.ReadFile(dataFile)
	if errRead != nil {
		return nil, fmt.Errorf("failed to read json file: %w", errRead)
	}

	// ураховуючи метод UnmarshalJSON(), додаємо поїзди до змінної result типу слайсу поїздів; обробляємо та повертаємо помилку, якщо така сталась.
	var res []Train
	errUnmarshal := json.Unmarshal(data, &res)

	if errUnmarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal json file: %w", errUnmarshal)
	}

	return res, nil
}

// повертає лише ті рейси, у яких станція відправлення та прибуття збігається з вхідними даними.
func getTrains(allTrains Trains, depStation, arrStation int) Trains {
	var result Trains

	for _, train := range allTrains {
		if train.DepartureStationID == depStation && train.ArrivalStationID == arrStation {
			result = append(result, train)
		}
	}

	return result
}

// сортування рейсів за критерієм за допомогою slicestable, оскільки він зберігає порядок.
func sortByCriteria(trains Trains, criteria string) Trains {
	switch criteria {
	case criteriaPrice:
		sort.SliceStable(trains, func(a, b int) bool {
			return trains[a].Price < trains[b].Price
		})
	case criteriaDepTime:
		sort.SliceStable(trains, func(a, b int) bool {
			return trains[b].DepartureTime.After(trains[a].DepartureTime)
		})
	case criteriaArrTime:
		sort.SliceStable(trains, func(a, b int) bool {
			return trains[b].ArrivalTime.After(trains[a].ArrivalTime)
		})
	}

	return trains
}

// виводить інформацію про кожний рейс у зручному для читання форматі.
func (trains Trains) printTrains() {
	for _, train := range trains {
		fmt.Printf(`Train ID: %v, Departure station ID: %v, Arrival station ID: %v, Price: %v, Departure time: %s, Arrival time: %s.%s`,
			train.TrainID,
			train.DepartureStationID,
			train.ArrivalStationID,
			train.Price,
			train.DepartureTime.Format(timeLayout),
			train.ArrivalTime.Format(timeLayout),
			"\n",
		)
	}
}
