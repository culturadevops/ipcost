package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var debug *log.Logger

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	debugEnv := os.Getenv("DEBUG")
	if debugEnv == "true" {
		debug = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)
	} else {
		debug = log.New(io.Discard, "", 0) // No-op logger
	}
}

func main() {

	ipCostsView()
}
func ExportToCSV(instances []EC2InstanceInfo, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"Region", "Name Tag", "Instance State", "Instance ID", "Public IP", "VPC ID", "Subnet ID", "Cost"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// Write instance data to CSV
	for _, instance := range instances {
		record := []string{
			instance.Region,
			instance.NameTag,
			instance.InstanceState,
			instance.InstanceID,
			instance.PublicIP,
			instance.VPCID,
			instance.SubnetID,
			fmt.Sprintf("%.2f", instance.Cost),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
		}
	}

	return nil
}

func ipCostsView() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	regions, err := fetchRegions(ec2Client)
	if err != nil {
		log.Fatalf("Failed to fetch regions: %v", err)
	}

	createAndPopulateInstancesTable(cfg, regions)
	//createAndPopulateLBTable(cfg, regions)
	return nil
}

func createAndPopulateInstancesTable(config aws.Config, regions []types.Region) error {

	debug.Println("Fetching all EC2 instances...")
	allInstances, err := fetchAllInstances(config, regions)
	if err != nil {
		debug.Printf("Error fetching all EC2 instances: %v", err)
		return nil
	}
	ExportToCSV(allInstances, "nuevoarchivo.xls")
	debug.Printf("Fetched %d EC2 instances.", len(allInstances))

	debug.Println("Sorting instances by IP...")

	debug.Println("Sorting done.")

	debug.Println("Populating table with instance data...")
	row := 1
	totalCost := 0.0
	for _, instanceInfo := range allInstances {
		fmt.Println(instanceInfo)
		totalCost += instanceInfo.Cost
		row++
	}
	debug.Println("Instances table population done.")

	debug.Printf("Total Instances IPs cost: $%.2f", totalCost)

	debug.Println("Finished createAndPopulateInstancesTable.")
	return nil
}
func createAndPopulateLBTable(cfg aws.Config, regions []types.Region) {
	debug.Println("Starting createAndPopulateLBTable...")
	debug.Println("Fetching all load balancers...")
	allLBs, err := fetchAllLoadBalancers(cfg, regions)
	if err != nil {
		debug.Printf("Error fetching all load balancers: %v", err)
	}
	debug.Printf("Fetched %d load balancers", len(allLBs))
	debug.Println("Sorting load balancers by IP...")
	row := 1
	totalIPCount := 0
	totalCost := 0.0
	debug.Println("Populating table with load balancer data...")
	for _, lbInfo := range allLBs {
		fmt.Println(lbInfo)
		row++

		totalIPCount += lbInfo.IPCount
		totalCost += lbInfo.Cost
	}
	debug.Printf("Finished createAndPopulateLBTable. Total IP Count: %d, Total Cost: %f", totalIPCount, totalCost)
}

func fetchRegions(client *ec2.Client) ([]types.Region, error) {
	regions, err := client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}
	return regions.Regions, nil
}
