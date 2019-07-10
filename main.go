package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/configservice"
	_ "github.com/motemen/go-loghttp/global"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"net/http"
	"os"
	"time"
)

var PROMETHEUS_ENDPOINT = os.Getenv("PROMETHEUS_ENDPOINT")

func main() {

	if PROMETHEUS_ENDPOINT == "" {
		log.Println("error: You must set the PROMETHEUS_ENDPOINT environment variable")
		os.Exit(1)
	}

		log.Println("info: starting app...")
	mux := http.NewServeMux()
	mux.HandleFunc("/AWSConfig", AWSConfigHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func AWSConfigHandler(w http.ResponseWriter, r *http.Request) {
	type Status struct {
		Status string `json:"status"`
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := AWSConfig(); err != nil {
		log.Println("error: Could not push metrics:", err)
		w.WriteHeader(http.StatusInternalServerError)
		s := Status{Status: "failed"}
		json.NewEncoder(w).Encode(s)
	} else {
		log.Println("info: metrics pushed successfully.")
		w.WriteHeader(http.StatusOK)
		s := Status{Status: "success"}
		json.NewEncoder(w).Encode(s)
	}

	return
}

func AWSConfig() error {

	ctx, cancelFn := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancelFn()

	svc := session.Must(session.NewSession())
	cs := configservice.New(svc, aws.NewConfig())
	params := &configservice.DescribeConfigRulesInput{}
	data, err := cs.DescribeConfigRulesWithContext(ctx, params)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, i := range data.ConfigRules {
		limit := int64(100)
		params := &configservice.GetComplianceDetailsByConfigRuleInput{ConfigRuleName: i.ConfigRuleName,
			Limit: &limit}
		config, err := cs.GetComplianceDetailsByConfigRuleWithContext(ctx, params)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, i := range config.EvaluationResults {
			time.Sleep(1 * time.Second)
			compliance := float64(1)
			log.Printf("info: %s %s %s\n", *i.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceId, *i.ComplianceType, *i.EvaluationResultIdentifier.EvaluationResultQualifier.ConfigRuleName)
			if *i.ComplianceType == "COMPLIANT" {
				compliance = 0
			}

			var (
				record = prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Name: "mfm_aws_config_compliance",
						Help: "mfm_aws_config_compliance AWS Config Compliance data",
					},
					[]string{"Resource", "ConfigRule", "ComplianceStatus"},
				)
			)

			pusher := push.New(PROMETHEUS_ENDPOINT, "mfm_aws_config_compliance")
			record.With(prometheus.Labels{
				"Resource":         *i.EvaluationResultIdentifier.EvaluationResultQualifier.ResourceId,
				"ConfigRule":       *i.EvaluationResultIdentifier.EvaluationResultQualifier.ConfigRuleName,
				"ComplianceStatus": *i.ComplianceType}).Set(compliance)

			pusher.Collector(record)
			if err := pusher.Push(); err != nil {
				log.Println("Could not push to Pushgateway:", err)
				return err
			}
		}
	}

	return nil

}
