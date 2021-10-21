package vconfig

import (
  "io/ioutil"
  "encoding/json"
)

type VerifierConfig struct {
  F int `json:"f"`
  Timer int `json:"timer"`
  Vip string `json:"verifierIP"`
}

func NewVerifierConfig() (*VerifierConfig, error) {
  vc := new(VerifierConfig)
  JSONBytes, _ := ioutil.ReadFile("verifier_config.json")
  err := json.Unmarshal(JSONBytes, vc)
  if err != nil {
    return nil, err
  }
  return vc, nil
}
