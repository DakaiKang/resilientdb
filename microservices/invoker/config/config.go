package config

import (
  // "fmt"
  "io/ioutil"
  "encoding/json"
)

type InvokerConfig struct {
  F int `json:"f"`
  ThresholdOn bool `json:"thresholdOn"`
  Threshold uint64 `json:"threshold"`
  Service string `json:"service"`
  FcnRegion string `json:"fcnRegion"`
  FcnName string `json:"fcnName"`
  Uuids [15]string `json:"uuids"`
}

func NewInvokerConfig() (*InvokerConfig, error) {
  ic := new(InvokerConfig)
  JSONBytes, _ := ioutil.ReadFile("invoker_config.json")
  err := json.Unmarshal(JSONBytes, ic)
  if err != nil {
    return nil, err
  }
  //fmt.Printf("CONFIG: %v\n", ic)
  return ic, nil
}
