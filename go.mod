module stressTest

go 1.21.1

require (
	github.com/gorilla/websocket v1.5.3
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/orcaman/concurrent-map/v2 v2.0.1
	gitlab.qa.dobest.com/qa/yokamancli v1.1.8
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.27.1

)

replace gitlab.qa.dobest.com/qa/yokamancli => ./yokamancli

require (
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)
