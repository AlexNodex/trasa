package integration_tests

import (
	"fmt"
	"github.com/seknox/trasa/server/api/crypt"
	"github.com/seknox/trasa/server/api/orgs"
	"github.com/seknox/trasa/server/api/stats"
	"github.com/seknox/trasa/server/api/system"
	"github.com/seknox/trasa/server/api/users"
	"github.com/seknox/trasa/server/global"
	"github.com/seknox/trasa/server/initdb"
	"io/ioutil"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("_________________________________________________________________________________________________________")
	setupTestEnv()
	m.Run()
}

func setupTestEnv() {
	testConfig := global.Config{
		Database: struct {
			Dbname     string `toml:"dbname"`
			Dbuser     string `toml:"dbuser"`
			Port       string `toml:"port"`
			Server     string `toml:"server"`
			Sslenabled bool   `toml:"sslenabled"`
			Usercert   string `toml:"usercert"`
			Userkey    string `toml:"userkey"`
			Cacert     string `toml:"cacert"`
		}{
			"trasadb",
			"trasauser",
			"54321",
			"127.0.0.1",
			false,
			"", "", "",
		},
		Logging: struct {
			Level         string `toml:"level"`
			SendErrReport string `toml:"sendErrReport"`
		}{"TRACE", ""},
		Minio: struct {
			Status bool   `toml:"status"`
			Key    string `toml:"key"`
			Secret string `toml:"secret"`
			Server string `toml:"server"`
			Usessl bool   `toml:"usessl"`
		}{false, "", "", "", false},
		Platform: struct {
			Base string `toml:"base"`
		}{"private"},
		Redis: struct {
			Port       string   `toml:"port"`
			Server     []string `toml:"server"`
			Sslenabled bool     `toml:"sslenabled"`
			Usercert   string   `toml:"usercert"`
			Userkey    string   `toml:"userkey"`
			Cacert     string   `toml:"cacert"`
		}{"", []string{"127.0.0.1:6379"}, false, "", "", ""},
		Timezone: struct {
			Location string `toml:"location"`
		}{"Asia/Kathmandu"},
		Security: struct {
			InsecureSkipVerify bool `toml:"insecureSkipVerify"`
		}{true},
		Trasa: struct {
			ListenAddr  string `toml:"listenAddr"`
			Dashboard   string `toml:"dashboard"`
			Rootdomain  string `toml:"rootdomain"`
			CloudServer string `toml:"cloudServer"`
			Ssodomain   string `toml:"ssodomain"`
			Trasacore   string `toml:"trasacore"`
			Rootdir     string `toml:"rootdir"`
			OrgId       string `toml:"orgID"`
		}{"localhost", "https://localhost", "", "https://u2fproxy.trasa.io", "", "", "", ""},
		SSHProxy: struct {
			ListenAddr string `toml:"listenAddr"`
		}{":8022"},
	}

	state := global.InitDBSTOREWithConfig(testConfig)
	err := insertMockData(state)
	if err != nil {
		panic(err)
	}

	crypt.InitStore(state)
	users.InitStore(state)
	orgs.InitStore(state)
	system.InitStore(state)
	stats.InitStore(state)

	initdb.InitDB()

}

func insertMockData(state *global.State) error {

	b, err := ioutil.ReadFile("mockdata.sql")
	if err != nil {
		return err
	}

	_, err2 := state.DB.Exec(string(b))
	if err2 != nil {
		fmt.Println(err2)
		return err2
	}

	return nil
}