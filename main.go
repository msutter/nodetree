package main

import (
	"io/ioutil"
	"nodetree/log"
	"nodetree/models"
	"os"
)

func main() {

	log.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	//  this are only test statements. Will be invoked by the future command line
	config := models.NewConfig()
	stage_tree := config.GetStageTree()

	log.Info.Println("")
	log.Info.Println("-------------- START SYNC LAB TREE --------------------")
	lab_stage := stage_tree.GetStageByName("lab")
	lab_stage.Sync()
	log.Info.Println("-------------- START SYNC LAB TREE --------------------")

	log.Info.Println("")
	log.Info.Println("-------------- START SYNC PRD TREE --------------------")
	prd_stage := stage_tree.GetStageByName("prd")
	prd_stage.Sync()
	log.Info.Println("-------------- START SYNC PRD TREE --------------------")

	log.Info.Println("")
	fqdns := []string{"pulp-lab-12411.local", "pulp-lab-11111.local"}
	log.Info.Printf("-------------- START SYNC FQDNS %v --------------------\n", fqdns)
	lab_stage.SyncByFilters(fqdns, []string{})
	log.Info.Printf("-------------- STOP SYNC FQDNS %v --------------------\n", fqdns)

	log.Info.Println("")
	tags := []string{"111MZ", "12MZ"}
	log.Info.Printf("-------------- START SYNC TAGS %v --------------------\n", tags)
	lab_stage.SyncByFilters([]string{}, tags)
	log.Info.Printf("-------------- STOP SYNC TAGS %v --------------------\n", tags)

	log.Info.Println("")
	log.Info.Printf("-------------- START SYNC BY Filters %v %v --------------------\n", fqdns, tags)
	lab_stage.SyncByFilters(fqdns, tags)
	log.Info.Printf("-------------- STOP SYNC BY Filters %v %v --------------------\n", fqdns, tags)
}
