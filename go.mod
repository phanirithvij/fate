module github.com/phanirithvij/fate

// +heroku goVersion go1.15
go 1.15

require (
	github.com/asdine/storm v2.1.2+incompatible
	github.com/fatih/color v1.10.0
	github.com/filebrowser/filebrowser/v2 v2.10.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.8.0
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0
	gorm.io/driver/postgres v1.0.5
	gorm.io/driver/sqlite v1.1.3
	gorm.io/gorm v1.20.7
)

replace github.com/filebrowser/filebrowser/v2 => github.com/phanirithvij/filebrowser/v2 v2.9.1-0.20201125121250-2d82696cf6bd
