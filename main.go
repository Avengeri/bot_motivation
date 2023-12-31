package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

// Глобальные переменные для того, чтобы впихнуть их во все щели
var gBot *tgbotapi.BotAPI

//var gChatId int64

var gUserInChat Users
var gUsefulActivities = Activities{
	{"Yoga", "Йога (15 минут)", 1},
	{"Book", "Чтение книги (15 минут)", 1},
	{"Walk", "Прогулка (15 минут)", 1},
	{"Sport", "Занятие спортом (15 минут)", 1},
	{"Pornhub", "Придушить змея (200 минут)", 10},
}
var gRewards = Activities{
	{"Film", "Посмотреть фильм", 5},
	{"Food", "Сожрать чего-нибудь вкусного", 7},
	{"Game", "Нагнуть нубов в контре", 10},
}
var chats = make(map[int64]*ChatState)

// Константы, чтобы код красивенький был, да и по уму как-то
const (
	EMOJI_COIN         = "\U0001FA99"   // (coin)
	EMOJI_SMILE        = "\U0001F642"   // 🙂
	EMOJI_SUNGLASSES   = "\U0001F60E"   // 😎
	EMOJI_WOW          = "\U0001F604"   // 😄
	EMOJI_DONT_KNOW    = "\U0001F937"   // 🤷
	EMOJI_SAD          = "\U0001F63F"   // 😿
	EMOJI_BICEPS       = "\U0001F4AA"   // 💪
	EMOJI_BUTTON_START = "\U000025B6  " // ▶
	EMOJI_BUTTON_END   = "  \U000025C0" // ◀

	BUTTON_TEXT_PRINT_INTRO       = EMOJI_BUTTON_START + "Посмотреть вступление" + EMOJI_BUTTON_END
	BUTTON_TEXT_SKIP_INTRO        = EMOJI_BUTTON_START + "Пропустить вступление" + EMOJI_BUTTON_END
	BUTTON_TEXT_BALANCE           = EMOJI_BUTTON_START + "Текущий баланс" + EMOJI_BUTTON_END
	BUTTON_TEXT_USEFUL_ACTIVITIES = EMOJI_BUTTON_START + "Полезные действия" + EMOJI_BUTTON_END
	BUTTON_TEXT_REWARDS           = EMOJI_BUTTON_START + "Награды" + EMOJI_BUTTON_END
	BUTTON_TEXT_PRINT_MENU        = EMOJI_BUTTON_START + "ОСНОВНОЕ МЕНЮ" + EMOJI_BUTTON_END

	BUTTON_CODE_PRINT_INTRO       = "print_intro"
	BUTTON_CODE_SKIP_INTRO        = "skip_intro"
	BUTTON_CODE_BALANCE           = "show_balance"
	BUTTON_CODE_USEFUL_ACTIVITIES = "show_useful_activities"
	BUTTON_CODE_REWARDS           = "show_rewards"
	BUTTON_CODE_PRINT_MENU        = "print_menu"

	UPDATE_CONFIG_TIMEOUT        = 60
	MAX_USER_COINS        uint16 = 500
)

type User struct {
	id    int
	name  string
	coins uint16
}

// Users не нужна для общения тет-а-тет, только для того, чтобы определить юзеров в чатах, где будет бот
type Users []*User

type Activity struct {
	code, name string
	coins      uint16
}
type Activities []*Activity
type ChatState struct {
	ChatID int64
	Users  map[int]*User
}

// Инициализация бота
func init() {
	//Загрузка, поиск и обработка ошибка токена
	envFilePath := "./go.env"

	err := godotenv.Load(envFilePath)
	if err != nil {
		log.Fatal("Не удалось загрузить переменную окружения")
	}
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Значение токена не установлено")
	} else {
		fmt.Printf("Значение токена: %s\n", botToken)
	}
	gBot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	gBot.Debug = true

	log.Printf("Authorized on account %s", gBot.Self.UserName)
}

// Старт
func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

// Проверка является ли данное обновление (update) обновлением типа CallbackQuery и содержит ли оно данные обратного вызова (callback data).
func isCallbackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}

// Задержка отправки сообщений
func delay(seconds uint8) {
	time.Sleep(time.Second * time.Duration(seconds))
}

// Сообщение с задержкой
func printSystemMessageWithDelay(chatID int64, delayInSec uint8, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	gBot.Send(msg)
	delay(delayInSec)
}

// Приветственное сообщение с задержкой
func printIntro(chatID int64, update *tgbotapi.Update) {
	printSystemMessageWithDelay(chatID, 2, "Привет!"+EMOJI_SUNGLASSES)
	printSystemMessageWithDelay(chatID, 2, "Этот бот поможет тебе быть замотивированным")
	printSystemMessageWithDelay(chatID, 2, "Выполняй полезные задачи, зарабатывай монетки и потом трать их")
	printSystemMessageWithDelay(chatID, 3, "Еще какая-нибудь шляпа о боте, но мне лень писать")
}

// Создание ряда кнопок встроенной клавиатуры
func getKeyboardRow(buttonText, buttonCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonCode))
}

// Выбор пропустить intro
func askToPrintIntro(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Во вступительном сообщение ты поймешь смысл этого бота, почитаем?")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_PRINT_INTRO, BUTTON_CODE_PRINT_INTRO),
		getKeyboardRow(BUTTON_TEXT_SKIP_INTRO, BUTTON_CODE_SKIP_INTRO),
	)
	gBot.Send(msg)
}

// Отображение меню
func showMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выбери один из вариантов")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_BALANCE, BUTTON_CODE_BALANCE),
		getKeyboardRow(BUTTON_TEXT_USEFUL_ACTIVITIES, BUTTON_CODE_USEFUL_ACTIVITIES),
		getKeyboardRow(BUTTON_TEXT_REWARDS, BUTTON_CODE_REWARDS),
	)
	gBot.Send(msg)
}

// Покажет баланс
func showBalance(chatID int64, user *User) {
	msg := fmt.Sprintf("%s, твой кошелек пока пуст (бомжара сраный) %s \nСделай чё нить полезное уже, и заработай, тряпка!", user.name, EMOJI_DONT_KNOW)
	if coins := user.coins; coins > 0 {
		msg = fmt.Sprintf("%s,у тебя %d %s", user.name, coins, EMOJI_COIN)
	}
	gBot.Send(tgbotapi.NewMessage(chatID, msg))

	showMenu(chatID)

}

// Служит для проверки, отсутствует ли информация о пользователе или данных обратного вызова (callback data) в объекте update
func callbackQueryIsMissing(update *tgbotapi.Update) bool {
	return update.CallbackQuery == nil || update.CallbackQuery.From == nil
}

// Получение юзера из обновления
func getUserFromUpdate(update *tgbotapi.Update) (user *User, found bool) {
	if callbackQueryIsMissing(update) {
		return
	}
	userId := update.CallbackQuery.From.ID
	for _, userInChat := range gUserInChat {
		if userId == userInChat.id {
			return userInChat, true
		}
	}
	return
}

// Извлекает информацию о пользователе из объекта update, проверяет, что событие типа CallbackQuery присутствует и не является нулевым, и затем сохраняет информацию о пользователе в глобальном списке gUserInChat
func storeUserFromUpdate(update *tgbotapi.Update, chatState *ChatState) (user *User, found bool) {
	if update.CallbackQuery == nil {
		return
	}

	from := update.CallbackQuery.From
	user = &User{id: from.ID, name: strings.TrimSpace(from.FirstName + " " + from.LastName), coins: 0}
	chatState.Users[user.id] = user
	return user, true
}

// Создает встроенную клавиатуру для бота в Telegram. Он берет список действий (activities), сообщение (message) и флаг isUseful, который указывает, полезное ли действие (если true, то отображает + перед монетами) или нет (если false, то отображает - перед монетами).
func showActivities(chatID int64, activities Activities, message string, isUseful bool) {
	activitiesButtonsRows := make([]([]tgbotapi.InlineKeyboardButton), 0, len(activities)+1)
	for _, activity := range activities {
		activityDescription := ""
		if isUseful {
			activityDescription = fmt.Sprintf("+ %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		} else {
			activityDescription = fmt.Sprintf("- %d %s: %s", activity.coins, EMOJI_COIN, activity.name)
		}
		activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(activityDescription, activity.code))
	}
	activitiesButtonsRows = append(activitiesButtonsRows, getKeyboardRow(BUTTON_TEXT_PRINT_MENU, BUTTON_CODE_PRINT_MENU))

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(activitiesButtonsRows...)
	gBot.Send(msg)
}

// Функция showUsefulActivities вызывает функцию showActivities для отображения полезных действий пользователю
func showUsefulActivities(chatID int64) {
	// Предполагается, что usefulActivities является глобальной переменной, содержащей полезные действия
	showActivities(chatID, gUsefulActivities, "Выберите полезное действие:", true)
}

// Функция showRewards вызывает функцию showActivities для отображения вознаграждений пользователю.
func showRewards(chatID int64) {
	// Предполагается, что rewards является глобальной переменной, содержащей награды
	showActivities(chatID, gRewards, "Выберите награду:", false)
}

// Используется для поиска действия в списке activities по заданному choiceCode. Она проходит по каждому элементу списка activities и сравнивает choiceCode с activity.code каждого действия. Если находит соответствие, то возвращает это действие (activity) и флаг true, указывая на успешное нахождение. Если не находит соответствия, возвращает nil и флаг false.
func findActivity(activities Activities, choiceCode string) (activity *Activity, found bool) {
	for _, activity := range activities {
		if choiceCode == activity.code {
			return activity, true
		}
	}
	return
}

// Предназначена для обработки полезного действия пользователя. Она принимает два аргумента: activity, представляющий информацию о действии, и user, представляющий информацию о пользователе.
func processUsefulActivity(activity *Activity, user *User, chatID int64) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`у активноси %s не указана стоимость`, activity.name)
	} else if user.coins+activity.coins > MAX_USER_COINS {
		errorMsg = fmt.Sprintf(`у тебя не может быть больше %d %s`, MAX_USER_COINS, EMOJI_COIN)
	}
	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf(`%s,прости, но %s %s Твой баланс остался без изменений`, user.name, errorMsg, EMOJI_SAD)
	} else {
		user.coins += activity.coins
		resultMessage = fmt.Sprintf(`%s, действие '%s' выполнено! %d %s поступило тебе на счет. Так держать! %s%s Теперь у тебя %d %s`,
			user.name, activity.name, activity.coins, EMOJI_COIN, EMOJI_BICEPS, EMOJI_SUNGLASSES, user.coins, EMOJI_COIN,
		)
		gBot.Send(tgbotapi.NewMessage(chatID, resultMessage))
	}
}

// Предназначена для обработки действия пользователя по получению награды. Она принимает два аргумента: activity, представляющий информацию о награде, и user, представляющий информацию о пользователе.
func processReward(activity *Activity, user *User, chatID int64) {
	errorMsg := ""
	if activity.coins == 0 {
		errorMsg = fmt.Sprintf(`у вознаграждения %s не указана стоимость`, activity.name)
	} else if user.coins < activity.coins {
		errorMsg = fmt.Sprintf(`у тебя сейчас %d %s. Ты не можешь себе позволить "%s" за %d %s`,
			user.coins, EMOJI_COIN, activity.name, activity.coins, EMOJI_COIN)
	}
	resultMessage := ""
	if errorMsg != "" {
		resultMessage = fmt.Sprintf("%s, прости, но %s %s твой баланс остался без изменений, вознаграждение недоустпно %s",
			user.name, errorMsg, EMOJI_SAD, EMOJI_DONT_KNOW)
	} else {
		user.coins -= activity.coins
		resultMessage = fmt.Sprintf(`%s, вознаграждение "%s" оплачено, приступай! %d %s было снято с твоего счета. Теперь у тебя %d %s`,
			user.name, activity.name, activity.coins, EMOJI_COIN, user.coins, EMOJI_COIN)
	}
	gBot.Send(tgbotapi.NewMessage(chatID, resultMessage))
}

// Обработка действий пользователя, связанных со встроенной клавиатурой бота в Telegram
func updateProcessing(update *tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	userID := update.CallbackQuery.From.ID

	chatState, exists := chats[chatID]
	if !exists {
		chatState = &ChatState{
			ChatID: chatID,
			Users:  make(map[int]*User),
		}
		chats[chatID] = chatState
	}

	user, exists := chatState.Users[userID]
	if !exists {
		user, _ = storeUserFromUpdate(update, chatState)
	}

	choiceCode := update.CallbackQuery.Data
	log.Printf("[%s] %s", time.Now(), choiceCode)

	switch choiceCode {
	case BUTTON_CODE_BALANCE:
		showBalance(chatID, user)
	case BUTTON_CODE_USEFUL_ACTIVITIES:
		showUsefulActivities(chatID)
	case BUTTON_CODE_REWARDS:
		showRewards(chatID)
	case BUTTON_CODE_PRINT_INTRO:
		printIntro(chatID, update)
		showMenu(chatID)
	case BUTTON_CODE_SKIP_INTRO:
		showMenu(chatID)
	case BUTTON_CODE_PRINT_MENU:
		showMenu(chatID)
	default:
		if usefulActivity, found := findActivity(gUsefulActivities, choiceCode); found {
			processUsefulActivity(usefulActivity, user, chatID)

			delay(2)
			showUsefulActivities(chatID)
			return
		}

		if reward, found := findActivity(gRewards, choiceCode); found {
			processReward(reward, user, chatID)

			delay(2)
			showRewards(chatID)
			return
		}
		log.Printf(`[%T]Неизвестный код "%s"!)`, time.Now(), choiceCode)
		msg := fmt.Sprintf("%s,прости,я не знаю код", user.name)
		gBot.Send(tgbotapi.NewMessage(chatID, msg))

	}
}
func startCommand(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	if _, exists := chats[chatID]; !exists {
		chats[chatID] = &ChatState{
			ChatID: chatID,
			Users:  make(map[int]*User),
		}
	}
}

func main() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	updates, err := gBot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal("Не удалось получить обновления с канала")
	}

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				gChatId := update.Message.Chat.ID
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				askToPrintIntro(gChatId)
			default:
				log.Printf("Неизвестная команда: %s", update.Message.Command())
			}
		} else if isCallbackQuery(&update) {
			updateProcessing(&update)
		}
	}
}
