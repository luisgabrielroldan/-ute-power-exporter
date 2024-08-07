package exporter

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	idbapi "github.com/influxdata/influxdb-client-go/v2/api"
	idbapiw "github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/rs/zerolog/log"

	"github.com/luisgabrielroldan/ute-power-exporter/internal/ute"
)

type InfluxDBClient interface {
	WriteAPIBlocking(org, bucket string) idbapi.WriteAPIBlocking
	Close()
}

type WriteAPIBlocking interface {
	WritePoint(ctx context.Context, p *idbapiw.Point) error
	Flush(ctx context.Context) error
}

type InfluxDBExporter struct {
	client InfluxDBClient
	org    string
	bucket string
}

type InfluxDBExporterConfig struct {
	ApiUrl string
	Org    string
	Bucket string
	Token  string
}

func NewInfluxDBExporter(config *InfluxDBExporterConfig) *InfluxDBExporter {
	client := influxdb2.NewClient(config.ApiUrl, config.Token)

	return &InfluxDBExporter{
		client: client,
		org:    config.Org,
		bucket: config.Bucket,
	}
}

func (e *InfluxDBExporter) Close() {
	e.client.Close()
}

func (e *InfluxDBExporter) Export(data *ute.ChargeCurveChartData) error {

	writeAPI := e.client.WriteAPIBlocking(e.org, e.bucket)

	for _, curve := range data.DataArray {
		log.Info().Msgf("Exporting curve %s", curve.Label)
		for _, point := range curve.Data {
			p := influxdb2.NewPointWithMeasurement("ute_power_consumption").
				AddTag("service", data.ServiceNumber).
				AddTag("label", curve.Label).
				AddField("value", point.Value).
				SetTime(time.Unix(point.Time/1000, 0))

			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}
	}

	err := writeAPI.Flush(context.Background())
	if err != nil {
		return err
	}

	return nil
}
