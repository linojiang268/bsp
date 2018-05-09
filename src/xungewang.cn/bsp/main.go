package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fvbock/endless"
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/go-ozzo/ozzo-routing"
	_ "github.com/lib/pq"
	"net/http"
	"strings"
	"time"
	"xungewang.cn/bsp/apis"
	"xungewang.cn/bsp/app"
	"xungewang.cn/bsp/repos"
	"xungewang.cn/bsp/services"
)

func main() {
	//parse command line to allow config directory to be specified via '-c path/to/config'
	configPath := flag.String("c", "", "the path(directory) where the config resides, e.g. ./config")
	flag.Parse()

	// load application-wide configuration
	var paths []string
	if len(*configPath) > 0 {
		paths = []string{*configPath}
	}
	if err := app.LoadConfig(paths...); err != nil {
		fmt.Printf("invalid application configuration: %s\n", err)
		fmt.Printf("you can either use -c to set path to config directory and make sure app.yaml in it with " +
			"expected properties or pass config from environment variables with prefix BSP_. e.g. BSP_HTTP_SERVER_ADDR, " +
			"BSP_DB_DSN, etc.")
		flag.Usage()

		return
	}

	setupLogger()
	db := setupDatabase()
	setupHttpServerAndStart(db)
}

func setupLogger() {
	// handle logging level
	if level, err := log.ParseLevel(app.Config.LogLevel); err == nil {
		log.SetLevel(level)
	}
}

func setupDatabase() *dbx.DB {
	log.Debugf("trying to connect to %s", app.Config.DSN)
	db, err := dbx.MustOpen("postgres", app.Config.DSN)
	if err != nil {
		log.Errorf("error while connecting to db: %s", err)
		panic(err)
	}
	log.Debugf("db connected")

	// set db's logging level to 'debug', by default there won't be
	// logging output(since the logging level is 'info' by default).
	// If you want to do things like showing SQL executed, lower the
	// default logging level.
	db.LogFunc = log.Debugf

	// config connection pool
	log.Debugf("max open conns: %d, max idle conns: %d, conn max lifetime: %d seconds",
		app.Config.DbMaxOpenConns, app.Config.DbMaxIdleConns, app.Config.DbConnMaxLifetime)

	db.DB().SetMaxOpenConns(app.Config.DbMaxOpenConns)
	db.DB().SetMaxIdleConns(app.Config.DbMaxIdleConns)
	db.DB().SetConnMaxLifetime(time.Duration(app.Config.DbConnMaxLifetime) * time.Second)

	return db
}

func setupHttpServerAndStart(db *dbx.DB) {
	// setup router
	http.Handle("/", setupRouter(db))

	// start server
	log.Infof("http server starts at %s", app.Config.HttpServerAddr)
	if err := endless.ListenAndServe(app.Config.HttpServerAddr, nil); err != nil {
		// sometimes it's fine since we may receive OS signals like SIGHUP. In that case, the error
		// message is 'use of closed network connection'.
		if !isClosedConnError(err) {
			fmt.Printf("failed to listen at %s, error: %s\n", app.Config.HttpServerAddr, err)
		}
	}
}

func setupRouter(db *dbx.DB) *routing.Router {
	router := routing.New()
	router.Use(app.Init())

	// api routers
	apis.SetupPositionRouter(router.Group("/api"), db, services.NewPositionService(repos.NewPositionRepo()))

	return router
}

// isClosedConnError reports whether err is an error from use of a closed
// network connection.
// copied from http2 package
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	// TODO: remove this string search and be more like the Windows
	// case below. That might involve modifying the standard library
	// to return better error types.
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}

	// DELIBERATELY COMMENT ACTIONS FOR WINDOWS FOR NOW
	//
	//// x/tools/cmd/bundle doesn't really support
	//// build tags, so I can't make an http2_windows.go file with
	//// Windows-specific stuff. Fix that and move this, once we
	//// have a way to bundle this into std's net/http somehow.
	//if runtime.GOOS == "windows" {
	//	if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
	//		if se, ok := oe.Err.(*os.SyscallError); ok && se.Syscall == "wsarecv" {
	//			const WSAECONNABORTED = 10053
	//			const WSAECONNRESET = 10054
	//			if n := errno(se.Err); n == WSAECONNRESET || n == WSAECONNABORTED {
	//				return true
	//			}
	//		}
	//	}
	//}
	return false
}
