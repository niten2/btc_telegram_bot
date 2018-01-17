package config

import (
  "os"
  "github.com/joho/godotenv"
  "strconv"
  "strings"
)

type Setting struct {
  DbUrl string
  DbName string
  TelegramToken string
  TelegramUserId string
  WalletId string
  ScheduleEverySeconds bool
}

func init() {
  _ = godotenv.Load()
}

func Settings() Setting {
  ScheduleEverySeconds, _ := strconv.ParseBool(os.Getenv("SCHEDULE_EVERY_SECONDS"))

  DbName := strings.Split(os.Getenv("DB_URL"), "/")[3]

  return Setting{
    DbUrl: os.Getenv("DB_URL"),
    DbName: DbName,
    TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
    TelegramUserId: os.Getenv("TELEGRAM_USER_ID"),
    WalletId: os.Getenv("WALLET_ID"),
    ScheduleEverySeconds: ScheduleEverySeconds,
  }
}