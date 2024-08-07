package main

import (
	"flag"
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
	"github.com/kouhin/envflag"
	"github.com/luisgabrielroldan/ute-power-exporter/internal/exporter"
	"github.com/luisgabrielroldan/ute-power-exporter/internal/ute"
)

type AppConfig struct {
	UteUsername      *string
	UtePassword      *string
	UteServiceNumber *string
	InfluxDBApiUrl   *string
	InfluxDBToken    *string
	InfluxDBOrg      *string
	InfluxDBBucket   *string
	FromDate         *string
	ToDate           *string
	Debug            *string
}

func parseConfig() *AppConfig {
	godotenv.Load()

	defaultFromDate := time.Now().AddDate(0, 0, -2).Format("02-01-2006")
	defaultToDate := time.Now().AddDate(0, 0, -1).Format("02-01-2006")

	config := &AppConfig{
		UteUsername:      flag.String("ute-user", "", "UTE Username"),
		UtePassword:      flag.String("ute-password", "", "UTE Password"),
		UteServiceNumber: flag.String("ute-service", "", "UTE Service Number"),
		InfluxDBApiUrl:   flag.String("influxdb-api-url", "", "InfluxDB API URL"),
		InfluxDBToken:    flag.String("influxdb-token", "", "InfluxDB Token"),
		InfluxDBOrg:      flag.String("influxdb-org", "", "InfluxDB Organization"),
		InfluxDBBucket:   flag.String("influxdb-bucket", "", "InfluxDB Bucket"),
		FromDate:         flag.String("from", defaultFromDate, "From date"),
		ToDate:           flag.String("to", defaultToDate, "To date"),
		Debug:            flag.String("debug", "false", "Enable debug log level"),
	}

	if err := envflag.Parse(); err != nil {
		panic(err)
	}

	return config
}

func configLogger(cfg *AppConfig) {
	if *cfg.Debug == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logOutput := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = logOutput.With().Logger()

	stdlog.SetFlags(0)
	stdlog.SetOutput(logOutput)
}

func main() {
	cfg := parseConfig()

	configLogger(cfg)

	fromDate, err := time.Parse("02-01-2006", *cfg.FromDate)
	if err != nil {
		log.Err(err).Msg("Invalid from date")
		return
	}

	toDate, err := time.Parse("02-01-2006", *cfg.ToDate)
	if err != nil {
		log.Err(err).Msg("Invalid to date")
		return
	}

	fromDateStr := fromDate.Format("02-01-2006")
	toDateStr := toDate.Format("02-01-2006")
	log.Info().Msgf("Getting charge curve data for service %s from %s to %s", *cfg.UteServiceNumber, fromDateStr, toDateStr)

	log.Info().Msg("Authorizing to UTE API...")

	client, err := ute.NewClient(*cfg.UteUsername, *cfg.UtePassword)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Info().Msg("Fetching charge curve data...")

	data, err := client.GetChargeCurveChartData(*cfg.UteServiceNumber, fromDate, toDate, ute.AllMagnitudes)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Info().Msg("Exporting data to InfluxDB...")

	influxExporter := exporter.NewInfluxDBExporter(&exporter.InfluxDBExporterConfig{
		ApiUrl: *cfg.InfluxDBApiUrl,
		Token:  *cfg.InfluxDBToken,
		Org:    *cfg.InfluxDBOrg,
		Bucket: *cfg.InfluxDBBucket,
	})

	defer influxExporter.Close()

	err = influxExporter.Export(data)
	if err != nil {
		log.Err(err).Msg("Error exporting data")
		return
	}

	log.Info().Msg("Export completed")
}
