package ute

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Magnitude string

const (
	ImportActiveEnergy Magnitude = "IMPORT_ACTIVE_ENERGY"
	ExportActiveEnergy Magnitude = "EXPORT_ACTIVE_ENERGY"
	Q1ReactiveEnergy   Magnitude = "Q1_REACTIVE_ENERGY"
	Q2ReactiveEnergy   Magnitude = "Q2_REACTIVE_ENERGY"
	Q3ReactiveEnergy   Magnitude = "Q3_REACTIVE_ENERGY"
	Q4ReactiveEnergy   Magnitude = "Q4_REACTIVE_ENERGY"
)

var AllMagnitudes = []Magnitude{ImportActiveEnergy, ExportActiveEnergy, Q1ReactiveEnergy, Q2ReactiveEnergy, Q3ReactiveEnergy, Q4ReactiveEnergy}

type GroupBy string

type Client struct {
	HTTPClient          *http.Client
	AutoServicioBaseUrl string
}

type DataPoint struct {
	Time  int64
	Value float64
}

func (dp *DataPoint) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return errors.New("invalid data point")
	}

	if err := json.Unmarshal(raw[0], &dp.Time); err != nil {
		return err
	}

	if err := json.Unmarshal(raw[1], &dp.Value); err != nil {
		return err
	}

	return nil
}

type ChargeCurve struct {
	Label string      `json:"label"`
	Data  []DataPoint `json:"data"`
}

type ChargeCurveChartData struct {
	ServiceNumber string
	DataArray     []ChargeCurve `json:"data_array"`
}

func (c *Client) GetChargeCurveChartData(serviceNumber string, startDate time.Time, endDate time.Time, magnitudes []Magnitude) (*ChargeCurveChartData, error) {
	baseURL := c.AutoServicioBaseUrl + "/SelfService/SSvcController/cmgraficarcurvadecarga"
	u, _ := url.Parse(baseURL)

	magnitudesStr := []string{}
	for _, m := range magnitudes {
		magnitudesStr = append(magnitudesStr, string(m))
	}

	query := url.Values{}
	query.Set("psId", serviceNumber)
	query.Set("fechaInicial", startDate.Format("02-01-2006"))
	query.Set("fechaFinal", endDate.Format("02-01-2006"))
	query.Set("agrupacion", "QH")
	query.Set("magnitudes", strings.Join(magnitudesStr, ","))
	u.RawQuery = query.Encode()

	apiUrl := u.String()

	res, err := c.HTTPClient.Get(apiUrl)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(string(data))
	}

	jsonStr := string(data)

	var respData ChargeCurveChartData

	err = json.Unmarshal([]byte(jsonStr), &respData)
	if err != nil {
		return nil, err
	}

	respData.ServiceNumber = serviceNumber

	return &respData, nil
}

func NewClient(username, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
	}

	if err := uteLogin(client, username, password); err != nil {
		return nil, err
	}

	if err := autoServicioLogin(client); err != nil {
		return nil, err
	}

	log.Debug().Msg("Client logged in successfully!")

	return &Client{
		HTTPClient:          client,
		AutoServicioBaseUrl: "https://autoservicio.ute.com.uy",
	}, nil
}
