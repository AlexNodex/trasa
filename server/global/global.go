package global

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
	"github.com/oschwald/geoip2-golang"

	firebase "firebase.google.com/go"
	"github.com/go-redis/redis/v8"
	vault "github.com/hashicorp/vault/api"

	//"go.etcd.io/etcd/clientv3"
	"github.com/seknox/trasa/server/models/migrations"
	"google.golang.org/api/option"
)

var DBVersion string = "2020-07-31-rc"

var config Config

func GetConfig() Config {
	return config
}

func SetOrgID(orgID string) {
	config.Trasa.OrgId = orgID
}

type State struct {
	DB             *sql.DB
	Geoip          *geoip2.Reader
	FirebaseClient *firebase.App
	MinioClient    *minio.Client
	VaultClient    *vault.Client
	//	EtcdClient     *clientv3.Client
	//Config      Config
	RedisClient    *redis.Client
	VaultRootToken string
	TsxvKey        tsxKey
	TsxCPxyKey     string
}

type tsxKey struct {
	Key   *[32]byte
	State bool
}

// ECDHKexDerivedKey stores kexDerivedKey with sessionID as key.
// For dashboard based login, key becomes sessionID of dashboard session.
// For enrol new device, key becomes trasaID of user
// For http session recording, key becomse sessionID of http session.
// It is responsibility of caller to delete keys after usage.
var ECDHKexDerivedKey = make(map[string]KexDerivedKey)

// KexDerivedKey stores secretkey derived from Kex, deviceID of device with which key was exchanged and timestamp recording kex operation.
type KexDerivedKey struct {
	// device id of client
	DeviceID string
	// secret key
	Secretkey []byte
	// Time of when secret key was derived
	Timestamp int64
}

func InitDBSTORE() *State {
	checkInitDirsAndFiles()
	conf := ParseConfig()
	return InitDBSTOREWithConfig(conf)
}

func InitDBSTOREWithConfig(conf Config) *State {

	config = conf
	level, _ := logrus.ParseLevel(config.Logging.Level)
	logOutputToFile := flag.Bool("f", false, "Write to file")

	flag.Parse()
	if *logOutputToFile {

		f, err := os.OpenFile("/var/log/trasa.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}

		logrus.SetOutput(f)
	} else {
		logrus.SetOutput(os.Stdout)
	}

	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	// we start trasa-server dependencies:

	// initialize cockroachdb connection
	db, err := sql.Open("postgres", DBconn(config))
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic("database connection down: " + err.Error())
	}

	//	DbCon = db

	redisClient := newRedisClient(config)

	//elasticUrl, authUser, authPass := elasticon()

	// initialize geoIP connection
	absPath, err := filepath.Abs("/etc/trasa/static/GeoLite2-City.mmdb")
	if err != nil {
		panic("geodb file not found: " + err.Error())
	}
	geodb, err := geoip2.Open(absPath)
	if err != nil {
		panic(err)
	}
	absPath, err = filepath.Abs("/etc/trasa/config/key.json")
	if err != nil {
		logrus.Errorf("firebase key not found: %v", err)
	}
	opt := option.WithCredentialsFile(absPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Errorf("firebase key not found: %v", err)
		//panic(err)
	}

	var minioClient *minio.Client
	if config.Minio.Status {
		minioClient, err = getMinioClient(config)
		if err != nil {
			panic(err)
		}
	}

	// DbEnv = &DBConn{
	// 	db:             db,
	// 	geoip:          geodb,
	// 	firebaseClient: app,
	// 	minioClient:    minioClient,

	// 	config:      config,
	// 	redisClient: redisClient,
	// }

	err = migrate(db)
	if err != nil {
		//fmt.Println(err)
		panic(err)
	}

	if config.Proxy.GuacdEnabled {
		guacdAddr := config.Proxy.GuacdAddr
		if guacdAddr == "" {
			guacdAddr = "127.0.0.1:4822"
		}
		c, err := net.Dial("tcp", guacdAddr)
		if err != nil {
			panic("guacd is down")
		}
		c.Close()
	}

	return &State{
		DB:             db,
		Geoip:          geodb,
		FirebaseClient: app,
		MinioClient:    minioClient,
		TsxvKey: tsxKey{
			Key:   new([32]byte),
			State: false,
		},
		RedisClient: redisClient,
	}

	//return

}

func migrate(conn *sql.DB) error {
	for _, v := range migrations.PrimaryMigration {
		_, e := conn.Exec(v)
		if e != nil {
			fmt.Println(e)
			return e

		}
		fmt.Printf("%s migrated\n", strings.Split(v, " ")[5])
	}
	return nil

}
func DBconn(config Config) string {

	dbuser := config.Database.Dbuser
	dbpass := config.Database.Dbpass
	dbhost := config.Database.Server
	dbport := config.Database.Port
	dbname := config.Database.Dbname

	sslEnabled := config.Database.Sslenabled
	var str string

	if sslEnabled {
		caCertPath := config.Database.Cacert
		userCertPath := config.Database.Usercert
		userKeyPath := config.Database.Userkey

		caCert, _ := filepath.Abs(caCertPath)
		userCert, _ := filepath.Abs(userCertPath)
		userKey, _ := filepath.Abs(userKeyPath)

		str = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=verify-full&sslrootcert=%s&sslcert=%s&sslkey=%s", dbuser, dbpass, dbhost, dbport, dbname, caCert, userCert, userKey)

	} else {
		str = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbuser, dbpass, dbhost, dbport, dbname)
	}

	return str

}

func getMinioClient(config Config) (*minio.Client, error) {

	endpoint := config.Minio.Server
	accessKeyID := config.Minio.Key
	secretAccessKey := config.Minio.Secret
	useSSL := config.Minio.Usessl
	insecureSkipVerify := config.Security.InsecureSkipVerify

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return nil, err
	}

	t := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	minioClient.SetCustomTransport(&t)

	//Check connection and if bucket exists
	bucketExists, err := minioClient.BucketExists("trasa-ssh-logs")
	if err != nil {
		panic(err)
	}
	if !bucketExists {
		err = minioClient.MakeBucket("trasa-ssh-logs", "")
		if err != nil {
			panic(err)
		}

	}
	bucketExists, err = minioClient.BucketExists("trasa-guac-logs")
	if err != nil {
		panic(err)
	}
	if !bucketExists {
		err = minioClient.MakeBucket("trasa-guac-logs", "")
		if err != nil {
			panic(err)
		}

	}
	bucketExists, err = minioClient.BucketExists("trasa-https-logs")
	if err != nil {
		panic(err)
	}
	if !bucketExists {
		err = minioClient.MakeBucket("trasa-https-logs", "")
		if err != nil {
			panic(err)
		}

	}
	bucketExists, err = minioClient.BucketExists("trasa-db-logs")
	if err != nil {
		panic(err)
	}
	if !bucketExists {
		err = minioClient.MakeBucket("trasa-db-logs", "")
		if err != nil {
			panic(err)
		}

	}

	return minioClient, nil

}

func newRedisClient(config Config) *redis.Client {

	cl := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Server[0],
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := cl.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return cl

}

func checkInitDirsAndFiles() {
	err := os.MkdirAll("/var/tmp/trasa/accessproxy/guac/shared", 0600)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/var/tmp/trasa/accessproxy/ssh", 0600)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/var/tmp/trasa/accessproxy/http", 0600)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("/var/trasa/crdb", 0600)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/var/trasa/minio", 0600)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("/etc/trasa/certs", 0600)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/etc/trasa/config", 0600)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/etc/trasa/static", 0600)
	if err != nil {
		panic(err)
	}

	//create config file if no exist
	_, err = os.Stat("/etc/trasa/config/config.toml")
	if err != nil {
		f, err := os.OpenFile("/etc/trasa/config/config.toml", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		f.WriteString(
			`[backup]
  backupdir = "$HOME/trasa/backup"

[database]
  cacert = "/etc/trasa/certs/app.trasa.io/ca.crt"
  dbname = "trasadb"
  dbtype = "postgres"
  dbuser = "trasauser"
  port = "5432"
  server = "localhost"
  usercert = "/etc/trasa/certs/app.trasa.io/client.trasauser.crt"
  userkey = "/etc/trasa/certs/app.trasa.io/client.trasauser.key"

[dbproxy]
  listenaddr = "127.0.0.1:8023"

[etcd]
  server = "http://localhost:2379"

[logging]
  level = "TRACE"

[minio]
  status = false
  key = "minioadmin"
  secret = "minioadmin"
  server = "127.0.0.1:9000"
  usessl = false

[platform]
  base = "private"

[redis]
  server = "localhost:6379"

[security]
  insecureSkipVerify = true

[proxy]
  sshlistenAddr = "127.0.0.1:8022"
  dbListenAddr = "127.0.0.1:3333"
  guacdEnabled = false
  guacdAddr = "127.0.0.1:4822"

[timezone]
  location = "Asia/Kathmandu"

[trasa]
  cloudserver = "https://sg.cpxy.trasa.io"
  dashboard = "http://localhost"
  listenaddr = "localhost"
  ssodomain = "sso.gw.trasa.io"
  orgID = ""

`)
	}

}
