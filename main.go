package main

import (
	"fmt"
	"time"

	"errors"
	"os"

	"math/rand"
	"strconv"

	"encoding/json"
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

func main() {
	//	... запит даних від користувача
	var departureStation, arrivalStation, criteria string

	fmt.Println("Please, specify your departure info:")

	fmt.Print("departureStation: ")
	fmt.Scanln(&departureStation)

	fmt.Print("arrivalStation: ")
	fmt.Scanln(&arrivalStation)

	fmt.Println(`!! -- Keep in mind: valid criteria values are "price", "arrival-time", "departure-time" (without quotes).`)
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

		//якщо час прибуття менший за час відправки, значить настав наступний день: переносимо відправлення на 31 грудня
		departureDay := train.DepartureTime.Day()
		departureMonth := train.DepartureTime.Month()
		if train.ArrivalTime.Before(train.DepartureTime) {
			departureDay = 31
			departureMonth = time.December
		}

		fmt.Printf(`Train ID: %v, Departure station ID: %v, Arrival station ID: %v, Price: %v, Departure time: "%v" of %s at %v hours %v minutes, Arrival time: "%v" of %s at %v hours %v minutes.`, 
			train.TrainID, 
			train.DepartureStationID, 
			train.ArrivalStationID, 
			train.Price,

			departureDay, 
			departureMonth, 
			train.DepartureTime.Hour(),
			train.DepartureTime.Minute(),

			train.ArrivalTime.Day(), 
			train.ArrivalTime.Month(), 
			train.ArrivalTime.Hour(),
			train.ArrivalTime.Minute(),
		)
		fmt.Println()
	}
}

//повертає 3 перші за сортуванням поїзди, що задовільняють станції прибуття та відправлення; перевіряє валідність вхідних даних користувача
func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	//обробка помилок з вхідних даних; конвертування айді станцій у інт значення
	departureStationID, arrivalStationID, err := checkInput(departureStation, arrivalStation, criteria)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// ... код
	// починаємо парсити файл джсон та обробляємо помилки, якщо виникають
	trains, errTrains := ParseJSON()
	if errTrains != nil {
		return nil, errTrains
	}

	//знаходимо усі рейси, що б відповідали заданим значенням depStation та arrStation
	results := getTrains(trains, departureStationID, arrivalStationID)

	//сортуємо варіанти за допомогою quicksort
	results = sortByCriteria(results, criteria)

	// маєте повернути перші 3 правильні значення (або ніл, якщо їх нема)
	switch len(results) {
	case 0: return nil, nil
	case 1, 2, 3: return results, nil
	default: return results[:3], nil 
	}

}

//перевіряє валідність вхідних даних, повертає числові значення станцій або помилку
func checkInput(depStation, arrStation, criteria string) (int, int, error) {
	// присвоєння значення та перевірка на валідність станції відправлення
	if depStation == "" {
		return 0, 0, errors.New("empty departure station")
	}
	depStationID, errDepStation := strconv.Atoi(depStation)
	if depStationID < 0 || errDepStation != nil {
		return 0, 0, errors.New("bad departure station input")
	}

	//присвоєння значення та перевірка на валідність станції прибуття
	if arrStation == "" {
		return 0, 0, errors.New("empty arrival station")
	}
	arrStationID, errArrStation := strconv.Atoi(arrStation)
	if arrStationID < 0 || errArrStation != nil {
		return 0, 0, errors.New("bad arrival station input")
	}

	//перевірка на валідний критерій для сортування
	if criteria != "price" && criteria != "arrival-time" && criteria != "departure-time" {
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

	//парсинг часу та обробка помилок
	parsedArrTime, errArrTime := time.Parse("15:04:05", res.ArrivalTime)
	if errArrTime != nil {
		return fmt.Errorf("failed to parse arrival time: %w", errArrTime)
	}
	parsedDepTime, errDepTime := time.Parse("15:04:05", res.DepartureTime)
	if errDepTime != nil {
		return fmt.Errorf("failed to parse departure time: %w", errDepTime)
	}

	//присвоюємо потрібній структурі значення з тимчасової, конвертуючи
	tr.ArrivalTime = time.Date(0, time.January, 1, parsedArrTime.Hour(), parsedArrTime.Minute(), parsedArrTime.Second(), 0, time.UTC)
	tr.DepartureTime = time.Date(0, time.January, 1, parsedDepTime.Hour(), parsedDepTime.Minute(), parsedDepTime.Second(), 0, time.UTC)

	//присвоюємо усі інші значення
	tr.TrainID = res.TrainID
	tr.DepartureStationID = res.DepartureStationID
	tr.ArrivalStationID = res.ArrivalStationID
	tr.Price = res.Price

	return nil //повернеться за відсутності помилок
}

//анмаршал джсон файла у трейнз
func ParseJSON() ([]Train, error) {
	//відкриваємо та читаємо файл, якщо не вийшло - повертаємо помилку
	data, errRead := ioutil.ReadFile("data.json")
	if errRead != nil {
		return nil, fmt.Errorf("failed to open json file: %w", errRead)
	}

	//ураховуючи метод "", додаємо поїди до змінної "результат" типу слайсу поїздів; обробляємо та повертаємо помилку, якщо така сталась
	var res []Train
	errUnmarshal := json.Unmarshal(data, &res)
	if errUnmarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal json file: %w", errUnmarshal)
	}

	return res, nil
}

//повертає лише ті рейси, у яких станція відправлення та прибуття збігається з вхідними даними
func getTrains(allTrains Trains, depStation int, arrStation int) Trains {
	var res Trains

	for _, train := range allTrains {
		if train.DepartureStationID == depStation && train.ArrivalStationID == arrStation {
			res = append(res, train)
		}
	}

	return res
}

//сортування рейсів за критерієм за допомогою quicksort
func sortByCriteria(trains Trains, criteria string) Trains {
	rand.Seed(time.Now().UnixNano())

	if len(trains) < 2 {
		return trains
	}

	left, right := 0, len(trains)-1
	pivot := rand.Int() % len(trains)
	trains[pivot], trains[right] = trains[right], trains[pivot]

	switch criteria {
	case "price":
		for i := 0; i < len(trains); i++ {
			if trains[i].Price < trains[right].Price {
				trains[left], trains[i] = trains[i], trains[left]
				left++
			}
		}
	case "departure-time":
		for i := 0; i < len(trains); i++ {
			if trains[right].DepartureTime.After(trains[i].DepartureTime) {
				trains[left], trains[i] = trains[i], trains[left]
		 		left++
			}
		}
	case "arrival-time":
		for i := 0; i < len(trains); i++ {
			if trains[right].ArrivalTime.After(trains[i].ArrivalTime) {
				trains[left], trains[i] = trains[i], trains[left]
		 		left++
			}
		}
	}

	trains[left], trains[right] = trains[right], trains[left]

	sortByCriteria(trains[:left], criteria)
    sortByCriteria(trains[left+1:], criteria)

	return trains
}