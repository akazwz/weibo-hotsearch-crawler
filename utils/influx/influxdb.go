package influx

import (
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
)

func Write(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) (err error) {
	client := influxdb2.NewClient(os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"))
	// always close client at the end
	defer client.Close()
	p := influxdb2.NewPoint(measurement, tags, fields, t)
	writeApi := client.WriteAPI(os.Getenv("INFLUXDB_ORG"), os.Getenv("INFLUXDB_BUCKET"))
	writeApi.WritePoint(p)
	writeApi.Flush()
	writeApi.SetWriteFailedCallback(func(batch string, error http.Error, retryAttempts uint) bool {
		err = &error
		return false
	})
	return
}
