package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/translate/v2"
	"io"
	"net/http"
	"os"
)

// Инициализация переменных окружения из файла .env в корневой директории проекта при запуске приложения
func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {

	// Retrieve environment variables

	myApp := app.New()
	myWindow := myApp.NewWindow("Приложение для погоды")

	// Создаем виджеты для ввода города и отображения результата
	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("Введите город")

	resultLabel := widget.NewLabel("")

	submitButton := widget.NewButton("Получить температуру", func() {
		// Получение температуры и обработка ошибок
		message, err := getTemperature(inputEntry.Text)
		if err != nil {
			resultLabel.SetText(err.Error())
			return
		}
		resultLabel.SetText("Температура в городе " + inputEntry.Text + ": " + message)
	})

	// Создаем контейнер и добавляем виджеты в контейнер в порядке отображения снизу вверх
	content := container.NewVBox(
		inputEntry,
		submitButton,
		// Добавляем пробелы слева и справа от результата для центрирования
		container.NewHBox(layout.NewSpacer(), resultLabel, layout.NewSpacer()),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()
}

func translateCity(russianCity string) (string, error) {
	// Получение API ключа из переменных окружения
	apiKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Google Translate API key not found")
	}

	// Инициализация клиента Google Translate API и контекста запроса к API
	ctx := context.Background()
	client, err := translate.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}

	// Вызов метода Translate API для перевода города с русского на английский язык
	resp, err := client.Translations.List([]string{russianCity}, "en").Do()
	if err != nil {
		return "", err
	}

	// Проверка наличия перевода города в ответе
	if len(resp.Translations) == 0 {
		return "", fmt.Errorf("Translation not found")
	}

	// Возвращение переведенного города
	return resp.Translations[0].TranslatedText, nil
}

func getTemperature(city string) (string, error) {
	// Получение API ключа из переменных окружения и перевод города на английский язык
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	city, err := translateCity(city)
	if err != nil {
		return "", err
	}
	if apiKey == "" {
		return "", fmt.Errorf("OpenWeatherMap API key not found")
	}

	// Формирование URL запроса к OpenWeatherMap API
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// закрытие тела ответа после завершения функции getTemperature
	defer resp.Body.Close()

	// Чтение тела ответа в переменную body типа []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	// Преобразование тела ответа в формате JSON в map[string]interface{} и проверка наличия ошибок
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Проверка наличия ошибок в ответе и наличия данных о температуре в ответе
	if (result["cod"] != nil && result["cod"] != 200.0) || result["main"] == nil {
		return "", fmt.Errorf("Город не найден")
	}

	temperature := result["main"].(map[string]interface{})["temp"]

	// Возвращение температуры в формате string с одним знаком после запятой
	return fmt.Sprintf("%.1f°C", temperature), nil
}
