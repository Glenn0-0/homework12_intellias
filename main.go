package main

import (
	"fmt"
	"time"
	"sort"
	"strconv"

	"errors" //пакети для обробки помилок
	"os"

	"encoding/json" //пакети для роботи з джсон
	"io/ioutil"
)

type Trains []Train

type Train struct {
	TrainID            int       `json: "trainId"`       
	DepartureStationID int       `json: "departureStationId"`      
	ArrivalStationID   int       `json: "arrivalStationId"`      
	Price              float32   `json: "price"`  
	ArrivalTime        time.Time `json: "arrivalTime"`
	DepartureTime      time.Time `json: "departureTime"`
}

const ( // константи критерій, адже ми їх перевикористовуємо
	criteriaPrice string = "price"
	criteriaDepTime = "departure-time"
	criteriaArrTime = "arrival-time"
)

func main() {
	//	... запит даних від користувача
	var departureStation, arrivalStation, criteria string

	fmt.Println("Please, specify your departure info:")

	fmt.Print("departureStation: ")
	fmt.Scanln(&departureStation)

	fmt.Print("arrivalStation: ")
	fmt.Scanln(&arrivalStation)

	fmt.Printf(`!! -- Keep in mind: valid criteria values are "%s", "%s", "%s" (without quotes).%s`, criteriaPrice, criteriaDepTime, criteriaArrTime, "\n")
	fmt.Print("criteria: ")
	fmt.Scanln(&criteria)

	//знаходимо 3 перші (за сортуванням) рейси, що б відповідали заданим значенням depStation та arrStation
	result, err := FindTrains(departureStation, arrivalStation, criteria)
	
	//	... обробка помилки
	if err != nil {
		fmt.Println(fmt.Errorf("operation failed: %w", err))
		os.Exit(0)
	}

	//	... друк result
	for _, train := range result {
		printTrain(train)
	}
}

//повертає 3 перші за сортуванням поїзди, що задовільняють станції прибуття та відправлення; перевіряє валідність вхідних даних користувача
func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	//обробка помилок з вхідних даних; у разі відсутності - конвертування айді станцій у інт значення
	departureStationID, arrivalStationID, err := checkInput(departureStation, arrivalStation, criteria)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// ... код
	// починаємо парсити файл джсон та обробляємо помилки, якщо виникають
	trains, errParse := parseJSON()
	if errParse != nil {
		return nil, errParse
	}

	//знаходимо усі рейси, що б відповідали заданим значенням depStation та arrStation
	results := getTrains(trains, departureStationID, arrivalStationID)

	//сортуємо варіанти за допомогою quicksort
	results = sortByCriteria(results, criteria)

	// маєте повернути перші 3 правильні значення (або ніл, якщо їх нема)
	switch len(results) {
	case 0: 
		return nil, nil
	case 1, 2: // світч-кейс щоб якщо результатів 0-2, програма не сварилась на індекс поза межами слайсу [:3]
		return results, nil 
	}

	return results[:3], nil

}

//перевіряє валідність вхідних даних, повертає числові значення станцій або помилку
func checkInput(depStation, arrStation, criteria string) (int, int, error) {
	// присвоєння значення та перевірка на валідність станції відправлення
	if depStation == "" {
		return 0, 0, errors.New("empty departure station")
	}
	depStationID, errDepStation := strconv.Atoi(depStation)
	if depStationID < 0 || errDepStation != nil { //перевірка на натуральне число
		return 0, 0, errors.New("bad departure station input")
	}

	//присвоєння значення та перевірка на валідність станції прибуття
	if arrStation == "" {
		return 0, 0, errors.New("empty arrival station")
	}
	arrStationID, errArrStation := strconv.Atoi(arrStation)
	if arrStationID < 0 || errArrStation != nil { //перевірка на натуральне число
		return 0, 0, errors.New("bad arrival station input")
	}

	//перевірка на валідний критерій для сортування
	if criteria != criteriaPrice && criteria != criteriaDepTime && criteria != criteriaArrTime {
		return 0, 0, errors.New("unsupported criteria")
	}

	return depStationID, arrStationID, nil
}

//метод для анмаршалінгу джсону, адже "гг:хх:сс" не парситься у поле типу тайм.Тайм
func (tr *Train) UnmarshalJSON(data []byte) error {
	type TrainStringTime struct {
		TrainID            int     `json: "trainId"`       
		DepartureStationID int     `json: "departureStationId"`      
		ArrivalStationID   int     `json: "arrivalStationId"`      
		Price              float32 `json: "price"`  
		ArrivalTime        string  `json: "arrivalTime"`
		DepartureTime      string  `json: "departureTime"`
	}

	//анмаршал у тимчасову структуру, зі зручним типом
	var res TrainStringTime
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	//парсинг часу відправлення та прибуття та обробка помилок
	parsedDepTime, errDepTime := time.Parse("15:04:05", res.DepartureTime)
	if errDepTime != nil {
		return fmt.Errorf("failed to parse departure time: %w", errDepTime)
	}
	parsedArrTime, errArrTime := time.Parse("15:04:05", res.ArrivalTime)
	if errArrTime != nil {
		return fmt.Errorf("failed to parse arrival time: %w", errArrTime)
	}

	//присвоюємо потрібній структурі значення з тимчасової, конвертуючи
	tr.DepartureTime = time.Date(1, time.January, 1, parsedDepTime.Hour(), parsedDepTime.Minute(), parsedDepTime.Second(), 0, time.UTC)
	tr.ArrivalTime = time.Date(1, time.January, 1, parsedArrTime.Hour(), parsedArrTime.Minute(), parsedArrTime.Second(), 0, time.UTC)

	//присвоюємо усі інші значення (без змін)
	tr.TrainID = res.TrainID
	tr.DepartureStationID = res.DepartureStationID
	tr.ArrivalStationID = res.ArrivalStationID
	tr.Price = res.Price

	return nil //повернеться за відсутності помилок
}

//анмаршал джсон файла у трейнз
func parseJSON() ([]Train, error) {
	//відкриваємо та читаємо файл, якщо не вийшло - повертаємо помилку
	data, errRead := ioutil.ReadFile("data.json")
	if errRead != nil {
		return nil, fmt.Errorf("failed to read json file: %w", errRead)
	}

	//ураховуючи метод UnmarshalJSON(), додаємо поїзди до змінної "результат" типу слайсу поїздів; обробляємо та повертаємо помилку, якщо така сталась
	var res []Train
	errUnmarshal := json.Unmarshal(data, &res)
	if errUnmarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal json file: %w", errUnmarshal)
	}

	return res, nil
}

//повертає лише ті рейси, у яких станція відправлення та прибуття збігається з вхідними даними
func getTrains(allTrains Trains, depStation int, arrStation int) Trains {
	var result Trains

	for _, train := range allTrains {
		if train.DepartureStationID == depStation && train.ArrivalStationID == arrStation {
			result = append(result, train)
		}
	}

	return result
}

//сортування рейсів за критерієм за допомогою slicestable, оскільки він зберігає порядок
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

//виводить інформацію про кожний рейс у зручному для читання форматі
func printTrain(tr Train) {
	fmt.Printf(`Train ID: %v, Departure station ID: %v, Arrival station ID: %v, Price: %v, Departure time: %s, Arrival time: %s.%s`, 
		tr.TrainID, 
		tr.DepartureStationID, 
		tr.ArrivalStationID, 
		tr.Price,
		tr.DepartureTime.Format("15:04:05"),
		tr.ArrivalTime.Format("15:04:05"),
		"\n",
	)
}