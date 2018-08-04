package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis"
	"github.com/joeshaw/envdecode"

	hmetrics "github.com/heroku/x/go-kit/metrics"
	"github.com/heroku/x/go-kit/metrics/provider/librato"
	log "github.com/sirupsen/logrus"
)

var provider hmetrics.Provider

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/img/", hashHandler)

	if err := http.ListenAndServe(bind, mux); err != nil {
		log.Panicln(err)
	}
}

func timer(label string, start time.Time) {
	log.WithFields(log.Fields{
		"label":    label,
		"duration": time.Since(start),
	}).Info("Measuring operation")
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	defer timer(r.URL.Path, time.Now())
	type data struct {
		ImageDataBase64 string
	}

	key := r.URL.Path
	var buf bytes.Buffer

	log.Println("About to load")
	if bs, ok := rc.Load(key); ok {
		_, err := buf.Write(bs)
		if err != nil {
			log.WithError(err)
		}
	} else {
		log.Println("Didn't get bytes")
	}

	if buf.Len() <= 0 {
		cacheCounter.With("miss").Add(1)
		log.Println("Gave up, recalculating")
		if err := createImage(strings.NewReader(key), &buf); err != nil {
			log.Panicln(err)
		}
		if err := rc.Store(key, buf.Bytes()); err != nil {
			log.Warn("Failed to store", err)
		}
	} else {
		cacheCounter.With("hit").Add(1)
	}

	var b64 bytes.Buffer

	encoder := base64.NewEncoder(base64.StdEncoding, &b64)

	_, err := encoder.Write(buf.Bytes())
	if err != nil {
		log.Panic(err)
	}

	if err = t.Execute(w, data{b64.String()}); err != nil {
		log.Panicln(err)
	}
}

func mustOpen(fileName string) {
	cmd := exec.Command("xdg-open", fileName)
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}
}

func createImage(r io.Reader, w io.Writer) error {
	h := sha256.New()
	io.Copy(h, r)

	sum := h.Sum(nil)
	bp := bitPNG{bytes: sum, mult: 16}

	b := time.Now()
	if err := png.Encode(w, bp); err != nil {
		return err
	}
	encodeHisto.Observe(float64(time.Since(b) / 1e6)) // Convert to ms

	return nil
}

// encodes a 128x128 image
type bitPNG struct {
	mult  int
	bytes []byte
}

func (img bitPNG) ColorModel() color.Model {
	return color.RGBAModel
}

func (img bitPNG) Bounds() image.Rectangle {
	return image.Rectangle{
		image.Point{0, 0},
		image.Point{1024, 1024},
	}
}

func (img bitPNG) At(x, y int) color.Color {
	x /= 1024 / img.mult
	y /= 1024 / img.mult
	pos := x + (y * 16)
	byt := pos / 8
	bit := pos % 8

	mask := byte(1) << byte(bit)

	on := (img.bytes[byt] & mask) > 0

	var r, g, b uint8 = 0, 0, 0
	if on {
		r, g, b = 128, 0, 128
	}

	// Random color shenanigans

	{
		rBit := (bit + 4) % 8
		mask = byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			r, b = 255, 255

		}
	}

	{
		rBit := (bit + 6) % 8
		mask = byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			b = 200
		}
	}

	{
		rBit := (bit + 2) % 8
		mask := byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			r = 64
		}
	}

	{
		rBit := (bit + 7) % 8
		mask := byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			g = 196
			b /= 2
		}
	}

	return color.RGBA{
		A: 255, R: r, G: g, B: b,
	}
}

var doc = `
	<!DOCTYPE html>
	<html>
	<head></head>
	<body>
		<img src="data:image/png;base64,{{ .ImageDataBase64 }}">
	</body>
	</html>
`

var t *template.Template

var bind string
var encodeHisto metrics.Histogram
var cacheCounter metrics.Counter

type config struct {
	LibratoUser     string `env:"LIBRATO_USER"`
	LibratoPassword string `env:"LIBRATO_TOKEN"`
	RedisURL        string `env:"REDIS_URL"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{})

	var err error
	t, err = template.New("output.html").Parse(doc)
	if err != nil {
		log.Panicln(err)
	}

	flag.StringVar(&bind, "bind", ":30000", "bind address")

	flag.Parse()

	var cfg config
	if err = envdecode.Decode(&cfg); err != nil {
		log.Warn(err)
	}

	libratoURL, err := url.Parse("https://metrics-api.librato.com/v1/measurements")
	if err != nil {
		log.Panic(err)
	}
	libratoURL.User = url.UserPassword(cfg.LibratoUser, cfg.LibratoPassword)

	provider = librato.New(libratoURL, time.Duration(30*time.Second))

	encodeHisto = provider.NewHistogram("encode.time", 3)
	cacheCounter = provider.NewCounter("cache")

	redisCfg, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Warn("Couldn't parse redis url")
		rc = noCache{}
	} else {
		rc = redisCache{redis.NewClient(redisCfg)}
	}
}

type cache interface {
	Load(key string) ([]byte, bool)
	Store(key string, bytes []byte) error
}

type redisCache struct {
	*redis.Client
}

func (rc redisCache) Load(key string) ([]byte, bool) {
	b, err := rc.Get(key).Bytes()
	if err != nil {
		return nil, false
	}

	return b, true
}

func (rc redisCache) Store(key string, bytes []byte) error {
	log.WithFields(log.Fields{
		"operation": "store",
		"key":       key,
		"bytes":     bytes,
	})
	return rc.Set(key, bytes, 0).Err()
}

var rc cache

type noCache struct {
}

func (c noCache) Load(key string) ([]byte, bool) {
	return nil, false
}

func (c noCache) Store(key string, bytes []byte) error {
	return nil
}
