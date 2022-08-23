package ugo

import (
	"crypto/rand"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/gomodule/redigo/redis"
	"github.com/joefazee/ugo/cache"
	"github.com/joefazee/ugo/mailer"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/joefazee/ugo/render"
	"github.com/joefazee/ugo/session"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	alphNum  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_+"
	version  = "1.0.0"
	envFile  = ".env"
	debug    = "DEBUG"
	port     = "PORT"
	renderer = "RENDERER"
)

var (
	redisCache  *cache.RedisCache
	badgerCache *cache.BadgerCache
	redisPool   *redis.Pool
	badgerConn  *badger.DB
)

type (
	Ugo struct {
		AppName       string
		Debug         bool
		Version       string
		ErrorLog      *log.Logger
		InfoLog       *log.Logger
		RootPath      string
		Routes        *chi.Mux
		config        config
		Render        *render.Render
		JetViews      *jet.Set
		Session       *scs.SessionManager
		DB            database
		EncryptionKey string
		Cache         cache.Cache
		Scheduler     *cron.Cron
		Mail          mailer.Mail
		Server        Server
	}

	Server struct {
		Name   string
		Port   string
		Secure bool
		URL    string
	}
	config struct {
		port        string
		renderer    string // Template rendering
		cookie      cookieConfig
		sessionType string
		database    databaseConfig
		redis       redisConfig
	}
)

func (u *Ugo) New(rootPath string) error {

	pathConfig := initPaths{
		rootPath: rootPath,
		folderNames: []string{
			"handlers",
			"migrations",
			"views",
			"mail",
			"data",
			"public",
			"tmp",
			"logs",
			"middleware",
		},
	}

	err := u.Init(pathConfig)

	if err != nil {
		return err
	}

	// Read and load e.nv
	err = u.checkDotEnv(rootPath)
	if err != nil {
		return err
	}
	err = godotenv.Load(rootPath + "/" + envFile)
	if err != nil {
		return err
	}

	// Init loggers
	infoLog, errorLog := u.createLoggers()

	//  init scheduler
	u.Scheduler = cron.New()

	// Connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := u.OpenDB(os.Getenv("DATABASE_TYPE"), u.BuildDSN())
		if err != nil {
			errorLog.Printf("error connecting to database: %s\n", err.Error())
			os.Exit(1)
		}

		u.DB = database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}

	}

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		redisCache = u.createClientRedisCache()
		u.Cache = redisCache
		redisPool = redisCache.Conn
	}

	if os.Getenv("CACHE") == "badger" || os.Getenv("SESSION_TYPE") == "badger" {
		badgerCache = u.createClientBadgerCache()
		u.Cache = badgerCache
		badgerConn = badgerCache.Conn

		_, err = u.Scheduler.AddFunc("@daily", func() {
			_ = badgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

	u.InfoLog = infoLog
	u.ErrorLog = errorLog
	isDebug, err := strconv.ParseBool(os.Getenv(debug))
	if err != nil {
		return err
	}

	u.Debug = isDebug
	u.Version = version
	u.RootPath = rootPath
	u.Mail = u.createMailer()
	u.config = config{
		port:     os.Getenv(port),
		renderer: os.Getenv(renderer),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      u.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	secure := true
	protocol := "https"
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
		protocol = "http"
	}

	u.Server = Server{
		Name:   os.Getenv("SERVER_NAME"),
		Port:   os.Getenv("PORT"),
		Secure: secure,
		URL:    fmt.Sprintf("%s://%s:%s", protocol, os.Getenv("SERVER_NAME"), os.Getenv("PORT")),
	}

	// inject session
	sess := session.Session{
		CookieLifetime: u.config.cookie.lifetime,
		CookiePersist:  u.config.cookie.persist,
		CookieName:     u.config.cookie.name,
		CookieDomain:   u.config.cookie.domain,
		SessionType:    u.config.sessionType,
		CookieSecure:   u.config.cookie.secure,
	}
	switch u.config.sessionType {
	case "redis":
		sess.RedisPool = redisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = u.DB.Pool
	}

	u.Session = sess.InitSession()
	u.EncryptionKey = os.Getenv("KEY")

	u.Routes = u.routes().(*chi.Mux)

	if u.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)
		u.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)
		u.JetViews = views
	}

	u.createRenderer()

	// start mailer
	go u.Mail.ListenForMail()

	return nil
}

func (u *Ugo) Init(p initPaths) error {

	root := p.rootPath
	for _, path := range p.folderNames {

		err := u.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return nil
		}
	}

	return nil
}

// ListenAndServe starts the web server
func (u *Ugo) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", u.Server.Name, u.config.port),
		ErrorLog:     u.ErrorLog,
		Handler:      u.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if u.DB.Pool != nil {
		defer u.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	u.InfoLog.Printf("Listening on %s:%s: Debug: %t\n", u.Server.Name, u.config.port, u.Debug)
	err := srv.ListenAndServe()

	if err != nil {
		u.ErrorLog.Panicln(err)
	}
}

func (u *Ugo) checkDotEnv(path string) error {

	err := u.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))

	if err != nil {
		return err
	}

	return nil
}

func (u *Ugo) createLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (u *Ugo) createRenderer() {

	u.Render = &render.Render{
		Renderer: u.config.renderer,
		RootPath: u.RootPath,
		Port:     u.config.port,
		JetViews: u.JetViews,
		Session:  u.Session,
	}

}

func (u *Ugo) createMailer() mailer.Mail {

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	return mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   u.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Result:      make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIURL:      os.Getenv("MAILER_URL"),
	}
}

func (u *Ugo) createClientRedisCache() *cache.RedisCache {
	return &cache.RedisCache{
		Conn:   u.createRedisPool(),
		Prefix: u.config.redis.prefix,
	}
}

func (u *Ugo) createClientBadgerCache() *cache.BadgerCache {
	return &cache.BadgerCache{
		Conn: u.createBadgerConn(),
	}
}

func (u *Ugo) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(u.RootPath + "tmp/badger"))
	if err != nil {
		u.ErrorLog.Println(err)
		return nil
	}
	return db
}

func (u *Ugo) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", u.config.redis.host, redis.DialPassword(u.config.redis.password))
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// BuildDSN build sql connection string
func (u *Ugo) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"),
		)

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	default:
		// i will be back

	}

	return dsn
}

func (u *Ugo) GenerateRandomString(n int) string {
	s, r := make([]rune, n), []rune(alphNum)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}
