package bot_api

import (
	"fmt"

	"gopkg.in/telegram-bot-api.v4"
	"os"
	"regexp"

	"app-telegram/logger"
	"github.com/sirupsen/logrus"

	"app-telegram/config"
	"app-telegram/models"
)

const MessageError = "Something went wrong"
const MessageUnknown = "The command is unknown, the list of command - help"
const MessageAddAlert = "Alert successfully added"
const MessageDataNotFound = "Data not found"

const MessageHelp = `
  possible commands \ n

   1. plist, список - list of altcoins and their values (polonex) \n
   2. an expression of the form "p SBD > 0.000020" sets the sending of the message by the bot if the following conditions occur \n
   3. settings, настройки - displays a description of the settings for the current user \n
   3. help, помощь - displays a description of the settings for the current user \n
`
const MessageSettings = `
  Here you will see a list of all the settings
`

func InitActions() {
	fmt.Println("InitActions for bot")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := Bot.GetUpdatesChan(u)

	if err != nil {
		logger.Log.Fatal(err)
	}

	// NOTE for testing
	if config.Settings().IsEnvTest {
		os.Exit(0)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		user_name := update.Message.From.UserName
		id_telegram := update.Message.Chat.ID
		text := update.Message.Text

		_, err = models.FindOrCreateUserByIdTelegram(user_name, id_telegram)

		if err != nil {
			logger.Log.Warn("FindOrCreateUserByIdTelegram", err)
		}

		logger.Log.WithFields(logrus.Fields{
			"user_name":   user_name,
			"text":        text,
			"id_telegram": id_telegram,
		}).Info("bot receive message")

		message := CreateResponse(text, id_telegram)
		response := tgbotapi.NewMessage(id_telegram, message)

		Bot.Send(response)
	}
}

func CreateResponse(input string, id_telegram int64) string {
	var msg string

	switch {
	case regexp.MustCompile(`^[p] [\D]* [\d\.]*`).MatchString(input):
		msg = CreateAlert(input, id_telegram)
	case regexp.MustCompile("(plist)|(список)").MatchString(input):
		msg = CreatePoloniexCoinList()
	case regexp.MustCompile("(help)|(помощь)|(помо)").MatchString(input):
		msg = MessageHelp
	case regexp.MustCompile("(settings)|(настройки)|(настр)").MatchString(input):
		msg = MessageSettings
	default:
		msg = MessageUnknown
	}

	return msg
}

func CreateAlert(input string, id_telegram int64) string {
	user, err := models.FindUserByIdTelegram(id_telegram)

	if err != nil {
		user = models.NewUser("name", id_telegram)
		user.Create()
	}

	alert, err := models.NewAlert(input)

	_, err = models.FindCoinByName(alert.Name)

	if err != nil {
		logger.Log.Warn(err)
		return MessageError
	}

	user.Alerts = append(user.Alerts, alert)
	err = user.Save()

	if err != nil {
		logger.Log.Warn(err)
		return MessageError
	}

	return MessageAddAlert
}

func CreatePoloniexCoinList() string {
	coins, err := models.FindCoinAll()

	if err != nil {
		logger.Log.Warn(err)
		return MessageError
	}

	if len(coins) == 0 {
		return MessageDataNotFound
	}

	return models.CreatePoloniexCoinList(coins)
}

func SendMessage(id_telegram int64, message string) {
	// NOTE for testing
	if config.Settings().IsEnvTest {
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"id_telegram": id_telegram,
		"message":     message,
	}).Info("bot send message")

	response := tgbotapi.NewMessage(id_telegram, message)
	Bot.Send(response)
}

// TODO add info btc by shedule
// func CreateInfoMessage() string {
//   usd := request.GetCoinmarketcapCurrentBtc()
//   balance_wallet := request.GetBitapsBalanceWallet()

//   return fmt.Sprintf(`
//     текущий курс %s
//     btc в кошельке 0.%d
//   `, usd, balance_wallet)
// }

// NOTE check coin
func CheckCoin() {
	logger.Log.Info("CheckCoin start")

	err := models.FetchCoin()

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"err": err,
		}).Warn("CheckCoin")
	} else {
		logger.Log.Info("FetchCoin successfully updated")
	}

	AddUsdtCoin()

	CheckUsersAlert()
}

func CheckUsersAlert() {
	users, err := models.FindUserAll()

	if err != nil {
		logger.Log.Warn(err)
	}

	for _, user := range users {
		message, err := user.CheckAndRemoveUserAlert()

		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"err": err,
			}).Warn("CheckUsersAlert")
		}

		if message != "" {
			SendMessage(user.IdTelegram, message)
		}
	}
}

func AddUsdtCoin() {
	coin, err := models.FindCoinByName("USDT_BTC")

	if err != nil {
		return
	}

	course_btc := coin.Value
	coins, err := models.FindCoinAll()

	for _, coin := range coins {
		coin.ValueUsdt = coin.Value * course_btc

		coin.Save()
	}
}
