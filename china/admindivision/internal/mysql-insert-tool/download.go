package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	sourceProvinceURL = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/provinces.json?raw=true"
	sourceCityURL     = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/cities.json?raw=true"
	sourceCountyURL   = "https://github.com/modood/Administrative-divisions-of-China/raw/master/dist/areas.json?raw=true"
	sourceTownURL     = "https://github.com/modood/Administrative-divisions-of-China/blob/master/dist/streets.json?raw=true"
	sourceVillageURL  = "https://github.com/modood/Administrative-divisions-of-China/blob/master/dist/villages.json?raw=true"
)

// 举例:
//
//	{"code":"110101001001","name":"多福巷社区居委会","streetCode":"110101001","areaCode":"110101","cityCode":"1101","provinceCode":"11"}
type adminNodeFormat struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Province string `json:"provinceCode"`
	City     string `json:"cityCode"`
	County   string `json:"areaCode"`
	Town     string `json:"streetCode"`
}

type downloadData struct {
	Provinces []adminNodeFormat
	Cities    []adminNodeFormat
	Counties  []adminNodeFormat
	Towns     []adminNodeFormat
	Villages  []adminNodeFormat
}

func downloadAdminDistricts() (data downloadData, _ error) {
	do := func(uri string, tgt *[]adminNodeFormat) error {
		data, err := downloadAndParse(uri)
		if err != nil {
			return fmt.Errorf("获取数据失败 (%w), URL '%s'", err, uri)
		}
		*tgt = data
		return nil
	}
	targets := []struct {
		uri string
		tgt *[]adminNodeFormat
	}{
		{sourceProvinceURL, &data.Provinces},
		{sourceCityURL, &data.Cities},
		{sourceCountyURL, &data.Counties},
		{sourceTownURL, &data.Towns},
		{sourceVillageURL, &data.Villages},
	}
	for _, t := range targets {
		if err := do(t.uri, t.tgt); err != nil {
			return data, err
		}
	}
	return data, nil
}

func downloadAndParse(uri string) (data []adminNodeFormat, _ error) {
	start := time.Now()
	cli := http.Client{Transport: http.DefaultTransport}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error (%w)", err)
	}
	rsp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cli.Do error (%w)", err)
	}
	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll error (%w)", err)
	}

	ela := time.Since(start)
	log.Printf("已下载数据量 %v, 耗时 %v", descSize(len(b)), ela)

	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error (%w)", err)
	}
	return data, nil
}

func descSize(s int) string {
	switch {
	case s < (1 << 10):
		return fmt.Sprintf("%dB", s)
	case s < (1 << 20):
		return fmt.Sprintf("%.1fKB", float32(s)/1024)
	case s < (1 << 30):
		return fmt.Sprintf("%.1fMB", float32(s)/1024/1024)
	default:
		return fmt.Sprintf("%.1fGB", float32(s)/1024/1024/1024)
	}
}
