package main

import (
	"fmt"
	"github.com/bcornu/gobernate/dao"
	"github.com/bcornu/gobernate/sample/model"
	"github.com/bcornu/gobernate/utils"
)

func main() {
	utils.InitDB(utils.DBConfig{
		Host:     "localhost",
		Port:     "5432",
		Database: "test",
		Login:    "test",
		Password: "testpassword"})

	missionDAO, err := dao.GetDAO(model.Mission{})
	if err != nil {
		panic(err)
	}
	// missionDAO.Log()

	_, err = missionDAO.Insert(model.Mission{Type: &model.MissionType{Name: "YOP"}})
	if err != nil {
		panic(err)
	}
	inserted, err := missionDAO.Insert(model.Mission{})
	newMission := inserted.(model.Mission)
	fmt.Println("inserted")
	fmt.Println(newMission)
	newMission.Active = true
	newMission.Name = "My new Mission Name"
	newMission.Type = &model.MissionType{Name: "YOP"}
	updated, err := missionDAO.Update(newMission)
	if err != nil {
		panic(err)
	}
	newMission = updated.(model.Mission)
	fmt.Println("updated")
	fmt.Println(newMission)

	mission, err := missionDAO.GetOne(int(newMission.Id))
	if err != nil {
		panic(err)
	}
	fmt.Println("GET ONE")
	fmt.Println(mission.(model.Mission))

	missions, err := missionDAO.GetAll()
	if err != nil {
		panic(err)
	}
	fmt.Println("GET ALL")
	fmt.Println(missions.([]model.Mission))
}
