module github.com/infobloxopen/atlas-app-toolkit

go 1.14

require (
	contrib.go.opencensus.io/exporter/ocagent v0.7.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/dgrijalva/jwt-go v3.2.1-0.20200107013213-dc14462fd587+incompatible
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.0
	github.com/lib/pq v1.3.1-0.20200116171513-9eb3fc897d6f
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.5.1
	go.opencensus.io v0.22.3
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	google.golang.org/api v0.26.0 // indirect
	google.golang.org/genproto v0.0.0-20200527145253-8367513e4ece
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200225123651-fc8f55426688
)
